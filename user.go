package doubanfm

import (
	"bytes"
	"mime/multipart"
)

type User struct {
	Id     string `json:"user_id"`
	Name   string `json:"user_name"`
	Email  string
	Token  string
	Expire string
}

func Login(uid, password string) (*User, error) {
	formdata := &bytes.Buffer{}
	w := multipart.NewWriter(formdata)
	w.WriteField("app_name", AppName)
	w.WriteField("version", AppVersion)
	w.WriteField("email", uid)
	w.WriteField("password", password)
	w.Close()

	resp, err := post(LoginUrl, w.FormDataContentType(), formdata)
	if err != nil {
		return nil, err
	}

	var r struct {
		User
		dfmError
	}

	if err = decode(resp, &r); err != nil {
		return nil, err
	}

	if r.R != 0 {
		return nil, &r.dfmError
	}
	return &r.User, nil
}
