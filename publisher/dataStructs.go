package publisher

type Request struct {
	id string
	OracleRequestFormat
}

type OracleRequestFormat struct {
	TYPE   string
	SOURCE string
	VALUE  string
}

type Response struct {
	REQUEST  string
	RESPOSNE map[string]string
}

func (r *Request) String() string {
	return "Sender: " + r.id + " Type: " + r.TYPE + " SOURCE: " + r.SOURCE + " VALUE: " + r.VALUE
}
