package entry

type Entry struct {
	ModuleName string `json:"moduleName,omitempty"`
	Host       string `json:"host,omitempty"`
	Event      string `json:"event,omitempty"`
	Level      string `json:"level,omitempty"`
	Time       string `json:"time,omitempty"`
	Request    []byte `json:"request,omitempty"`
	Response   []byte `json:"response,omitempty"`
	ErrorText  string `json:"errorText,omitempty"`
}
