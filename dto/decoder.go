package dto

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/pkg/errors"
)

type Decoder struct {
	decoder *schema.Decoder
}

func NewDecoder() *Decoder {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	decoder.SetAliasTag("json")
	return &Decoder{decoder: decoder}
}

func (d *Decoder) Decode(payload interface{}, r *http.Request) error {
	if r.Method == http.MethodPost {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			return errors.Wrapf(err, "cannot read body of http request")
		}
		err = json.Unmarshal(buf, payload)
		if err != nil {
			return errors.Wrap(err, "cannot json unmarshal")
		}
		return nil
	}
	if m, ok := payload.(Payload); ok {
		for k, v := range r.URL.Query() {
			if len(v) > 1 {
				m[k] = v
			} else if len(v) == 0 {
				m[k] = ""
			} else {
				m[k] = v[0]
			}
		}
		return nil
	}

	err := d.decoder.Decode(payload, r.URL.Query())
	if err != nil {
		return errors.Wrap(err, "fails to decode")
	}
	return nil
}
