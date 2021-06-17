package response

/*
	response package providers information about returning data from server.
*/

type Response struct {
	Error string `json:"error"`
	ID    string `json:"id"`
}

type List struct {
	ID  string `json:"id"`
	TTL int    `json:"ttl"`
}
