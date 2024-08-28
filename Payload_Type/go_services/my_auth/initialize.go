package my_auth

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/MythicMeta/MythicContainer/authstructs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/utils/sharedStructs"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var samlSP *samlsp.Middleware

func loadConfig() (map[string]string, error) {
	configBytes, err := os.ReadFile(filepath.Join(".", "my_auth_config.json"))
	if err != nil {
		return nil, err
	}
	currentConfig := make(map[string]string)
	err = json.Unmarshal(configBytes, &currentConfig)
	if err != nil {
		logging.LogError(err, "Failed to unmarshal config bytes")
		return nil, err
	}
	return currentConfig, nil
}
func initializeSAMLSP(authName string, serverName string) error {
	if samlSP != nil {
		return nil
	}
	err := generateCerts(authName)
	if err != nil {
		return err
	}
	keyPair, err := tls.LoadX509KeyPair(fmt.Sprintf("%s.crt", authName),
		fmt.Sprintf("%s.key", authName))
	if err != nil {
		return err
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return err
	}
	config, err := loadConfig()
	if err != nil {
		return err
	}
	idpFileMetadata := config["metadata_file"]
	idpFileBytes, err := os.ReadFile(idpFileMetadata)
	if err != nil {
		return err
	}
	idpMetadata := saml.EntityDescriptor{}
	err = xml.Unmarshal(idpFileBytes, &idpMetadata)
	if err != nil {
		return err
	}
	rootURL, err := url.Parse(fmt.Sprintf("https://%s:7443", serverName))
	if err != nil {
		return err
	}
	acsURL, err := url.Parse(fmt.Sprintf("https://%s:7443/auth_acs/%s/ADFS", serverName, authName))
	if err != nil {
		return err
	}
	samlSP, _ = samlsp.New(samlsp.Options{
		URL:                 *rootURL,
		Key:                 keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate:         keyPair.Leaf,
		IDPMetadata:         &idpMetadata,
		UseArtifactResponse: false,
	})
	samlSP.ServiceProvider.AcsURL = *acsURL
	samlSP.ServiceProvider.MetadataValidDuration = 365 * 24 * time.Hour
	samlSP.ServiceProvider.EntityID = rootURL.String()
	samlSP.ServiceProvider.AuthnNameIDFormat = saml.UnspecifiedNameIDFormat // needed for ADFS
	return nil
}

/*
`my_auth_config` is used to determine the name of the file that has the ADFS metadata to use for SSO.
`MyAuthProvider.crt` and `MyAuthProvider.key` are generated and used for ADFS as well.
*/
func Initialize() {
	authName := "MyAuthProvider"
	myAuth := authstructs.AuthDefinition{
		Name:           authName,
		Description:    "A custom SSO auth provider for ADFS",
		IDPServices:    []string{"ADFS"},
		NonIDPServices: []string{"LDAP"},
		OnContainerStartFunction: func(message sharedStructs.ContainerOnStartMessage) sharedStructs.ContainerOnStartMessageResponse {
			logging.LogInfo("started", "inputMsg", message)
			return sharedStructs.ContainerOnStartMessageResponse{}
		},
		GetIDPMetadata: func(message authstructs.GetIDPMetadataMessage) authstructs.GetIDPMetadataMessageResponse {
			response := authstructs.GetIDPMetadataMessageResponse{
				Success: false,
			}
			err := initializeSAMLSP(authName, message.ServerName)
			if err != nil {
				response.Error = err.Error()
				return response
			}
			buf, err := xml.MarshalIndent(samlSP.ServiceProvider.Metadata(), "", " ")
			if err != nil {
				response.Error = err.Error()
				return response
			}
			response.Success = true
			response.Metadata = string(buf)
			return response
		},
		// https://github.com/crewjam/saml/blob/main/samlsp/middleware.go#L127
		GetIDPRedirect: func(message authstructs.GetIDPRedirectMessage) authstructs.GetIDPRedirectMessageResponse {
			response := authstructs.GetIDPRedirectMessageResponse{
				Success: false,
			}
			err := initializeSAMLSP(authName, message.ServerName)
			if err != nil {
				response.Error = err.Error()
				return response
			}
			binding := saml.HTTPRedirectBinding
			bindingLocation := samlSP.ServiceProvider.GetSSOBindingLocation(saml.HTTPRedirectBinding)
			authReq, err := samlSP.ServiceProvider.MakeAuthenticationRequest(bindingLocation, binding, saml.HTTPPostBinding)
			if err != nil {
				response.Error = err.Error()
				return response
			}
			newReq, err := http.NewRequest(http.MethodGet, message.RequestURL, nil)
			if err != nil {
				response.Error = err.Error()
				return response
			}
			for key, value := range message.RequestHeaders {
				newReq.Header.Add(key, value)
			}
			cookie := ""
			for _, value := range message.RequestCookies {
				cookie += value + ";"
			}
			if cookie != "" {
				newReq.Header.Add("Cookie", cookie)
			}
			for key, value := range message.RequestQuery {
				newReq.URL.Query().Add(key, value)
			}
			recorder := httptest.NewRecorder()
			relayState, err := samlSP.RequestTracker.TrackRequest(recorder, newReq, authReq.ID)
			if err != nil {
				response.Error = err.Error()
				return response
			}
			redirectURL, err := authReq.Redirect(relayState, &samlSP.ServiceProvider)
			if err != nil {
				response.Error = err.Error()
				return response
			}
			cookies := make(map[string]string)
			for _, resultCookie := range recorder.Result().Cookies() {
				cookies[resultCookie.Name] = resultCookie.String()
			}
			headers := make(map[string]string)
			for key, value := range recorder.Result().Header {
				headers[key] = strings.Join(value, ",")
			}
			response.Success = true
			response.RedirectURL = redirectURL.String()
			response.RedirectHeaders = headers
			response.RedirectCookies = cookies
			return response
		},
		ProcessIDPResponse: func(message authstructs.ProcessIDPResponseMessage) authstructs.ProcessIDPResponseMessageResponse {
			response := authstructs.ProcessIDPResponseMessageResponse{
				SuccessfulAuthentication: false,
			}
			err := initializeSAMLSP(authName, message.ServerName)
			if err != nil {
				response.Error = err.Error()
				return response
			}
			newReq, err := http.NewRequest(http.MethodPost, message.RequestURL, strings.NewReader(message.RequestBody))
			if err != nil {
				response.Error = err.Error()
				return response
			}
			for key, value := range message.RequestHeaders {
				newReq.Header.Add(key, value)
			}
			cookie := ""
			for _, value := range message.RequestCookies {
				cookie += value + ";"
			}
			if cookie != "" {
				newReq.Header.Add("Cookie", cookie)
			}
			for key, value := range message.RequestQuery {
				newReq.URL.Query().Add(key, value)
			}
			err = newReq.ParseForm()
			if err != nil {
				response.Error = err.Error()
				return response
			}
			possibleRequestIDs := []string{}
			if samlSP.ServiceProvider.AllowIDPInitiated {
				possibleRequestIDs = append(possibleRequestIDs, "")
			}
			trackedRequests := samlSP.RequestTracker.GetTrackedRequests(newReq)
			for _, tr := range trackedRequests {
				possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
			}
			assertion, err := samlSP.ServiceProvider.ParseResponse(newReq, possibleRequestIDs)
			if err != nil {
				if authErr, ok := err.(*saml.InvalidResponseError); ok {
					logging.LogError(err, fmt.Sprintf("failed to validate SAML response: %v", authErr.PrivateErr))
				}
				response.Error = err.Error()
				return response
			}
			logging.LogInfo("parsed assertions", "assertion", assertion)
			redirectURI := "/new/callbacks"
			if trackedRequestIndex := newReq.Form.Get("RelayState"); trackedRequestIndex != "" {
				trackedRequest, err := samlSP.RequestTracker.GetTrackedRequest(newReq, trackedRequestIndex)
				if err != nil {
					if errors.Is(err, http.ErrNoCookie) && samlSP.ServiceProvider.AllowIDPInitiated {
						if uri := newReq.Form.Get("RelayState"); uri != "" {
							redirectURI = uri
						}
					} else {
						response.Error = err.Error()
						return response
					}
				} else {
					err = samlSP.RequestTracker.StopTrackingRequest(httptest.NewRecorder(), newReq, trackedRequestIndex)
					if err != nil {
						response.Error = err.Error()
						return response
					}
					redirectURI = trackedRequest.URI
				}
			}
			recorder := httptest.NewRecorder()
			logging.LogInfo("potential redirect ui", "redirectURI", redirectURI)
			err = samlSP.Session.CreateSession(recorder, newReq, assertion)
			if err != nil {
				response.Error = err.Error()
				return response
			}
			cookies := make(map[string]string)
			for _, resultCookie := range recorder.Result().Cookies() {
				cookies[resultCookie.Name] = resultCookie.String()
			}
			headers := make(map[string]string)
			for key, value := range recorder.Result().Header {
				headers[key] = strings.Join(value, ",")
			}
			response.SuccessfulAuthentication = true
			if len(assertion.AttributeStatements) > 0 {
				if len(assertion.AttributeStatements[0].Attributes) > 0 {
					if len(assertion.AttributeStatements[0].Attributes[0].Values) > 0 {
						response.Email = assertion.AttributeStatements[0].Attributes[0].Values[0].Value
					}
				}
			}
			logging.LogInfo("result cookies", "cookies", cookies)
			logging.LogInfo("result headers", "headers", headers)
			return response
		},
		GetNonIDPMetadata: func(message authstructs.GetNonIDPMetadataMessage) authstructs.GetNonIDPMetadataMessageResponse {
			return authstructs.GetNonIDPMetadataMessageResponse{
				Success:  true,
				Metadata: "Requires Username, Password, and OTP code for successful auth via LDAP",
			}
		},
		GetNonIDPRedirect: func(message authstructs.GetNonIDPRedirectMessage) authstructs.GetNonIDPRedirectMessageResponse {
			return authstructs.GetNonIDPRedirectMessageResponse{
				Success:       true,
				RequestFields: []string{"username", "password", "OTP"},
			}
		},
		ProcessNonIDPResponse: func(message authstructs.ProcessNonIDPResponseMessage) authstructs.ProcessNonIDPResponseMessageResponse {
			return authstructs.ProcessNonIDPResponseMessageResponse{
				SuccessfulAuthentication: true,
				Email:                    message.RequestValues["username"],
			}
		},
	}
	authstructs.AllAuthData.Get(authName).AddAuthDefinition(myAuth)
}
