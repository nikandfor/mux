package mux

type (
	Responder interface {
		Respond(code int, msg interface{}) error
	}

	Encoder interface {
		Encode(msg interface{}) error
	}

	ResponderFunc func(code int, msg interface{}) error
	EncoderFunc   func(msg interface{}) error
)

func (f ResponderFunc) Respond(code int, msg interface{}) error { return f(code, msg) }
func (f EncoderFunc) Encode(msg interface{}) error              { return f(msg) }
