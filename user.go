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

func Login(id, password string) (*User, error) {
	formdata := &bytes.Buffer{}

	w := multipart.NewWriter(formdata)
	w.WriteField("app_name", AppName)
	w.WriteField("version", AppVersion)
	w.WriteField("email", id)
	w.WriteField("password", password)
	defer w.Close()

	resp, err := post(LoginUrl, w.FormDataContentType(), formdata)
	if err != nil {
		return nil, err
	}

	var r struct {
		User
		dbError
	}

	if err = decode(resp, &r); err != nil {
		return nil, err
	}

	if r.R != 0 {
		return nil, &r.dbError
	}
	return &r.User, nil
}
