// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package entry

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson5f4debdaDecodeGithubComIntegrationSystemIspJournalEntry(in *jlexer.Lexer, out *Entry) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "moduleName":
			out.ModuleName = string(in.String())
		case "host":
			out.Host = string(in.String())
		case "event":
			out.Event = string(in.String())
		case "level":
			out.Level = string(in.String())
		case "time":
			out.Time = string(in.String())
		case "request":
			if in.IsNull() {
				in.Skip()
				out.Request = nil
			} else {
				out.Request = in.Bytes()
			}
		case "response":
			if in.IsNull() {
				in.Skip()
				out.Response = nil
			} else {
				out.Response = in.Bytes()
			}
		case "errorText":
			out.ErrorText = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson5f4debdaEncodeGithubComIntegrationSystemIspJournalEntry(out *jwriter.Writer, in Entry) {
	out.RawByte('{')
	first := true
	_ = first
	if in.ModuleName != "" {
		const prefix string = ",\"moduleName\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.ModuleName))
	}
	if in.Host != "" {
		const prefix string = ",\"host\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Host))
	}
	if in.Event != "" {
		const prefix string = ",\"event\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Event))
	}
	if in.Level != "" {
		const prefix string = ",\"level\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Level))
	}
	if in.Time != "" {
		const prefix string = ",\"time\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Time))
	}
	if len(in.Request) != 0 {
		const prefix string = ",\"request\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Base64Bytes(in.Request)
	}
	if len(in.Response) != 0 {
		const prefix string = ",\"response\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Base64Bytes(in.Response)
	}
	if in.ErrorText != "" {
		const prefix string = ",\"errorText\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.ErrorText))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Entry) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson5f4debdaEncodeGithubComIntegrationSystemIspJournalEntry(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Entry) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson5f4debdaEncodeGithubComIntegrationSystemIspJournalEntry(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Entry) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson5f4debdaDecodeGithubComIntegrationSystemIspJournalEntry(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Entry) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson5f4debdaDecodeGithubComIntegrationSystemIspJournalEntry(l, v)
}