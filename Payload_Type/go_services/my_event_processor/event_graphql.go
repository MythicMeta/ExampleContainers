package my_event_processor

const CreateTagType = `# @genqlient
  mutation CreateNewTagType($color: String!, $description: String!, $name: String!) {
      insert_tagtype_one(object: {color: $color, description: $description, name: $name}, on_conflict: {constraint: tagtype_name_operation_id_key, update_columns: color}) {
        id
      }
    }
`
const CreateNewTagMutation = `# @genqlient
	mutation CreateNewTag($tagtype_id: Int!, $source: String!, $url: String!, $data: jsonb!, $filemeta_id: Int!) {
	  insert_tag_one(object: {data: $data, source: $source, tagtype_id: $tagtype_id, url: $url, filemeta_id: $filemeta_id}) {
		id
	  }
	}
`
const GetPayloadDataQuery = `# @genqlient
  query GetPayloadData($uuid: String!) {
      payload(where: {uuid: {_eq: $uuid}}) {
        filemetum{
			id
		}
      }
    }
`
