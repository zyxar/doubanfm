package doubanfm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"time"
)

type User struct {
	Id     string `json:"user_id"`
	Name   string `json:"user_name"`
	Email  string
	Token  string
	Expire string
}

func (this User) String() string {
	return fmt.Sprintf("\r    Id:\t%s\n  Name:\t%s\n Token:\t%s\nExpire:\t%s",
		this.Id, this.Name, this.Token, parseTime(this.Expire))
}

func (this User) Json() string {
	v, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(v)
}

func (this User) Save(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(this)
}

func (this User) SaveFile(fn string) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	return this.Save(f)
}

func (this *User) Load(r io.Reader) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(this)
}

func (this *User) LoadFile(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	return this.Load(f)
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

	if err = json.NewDecoder(resp).Decode(&r); err != nil {
		return nil, err
	}

	if r.R != 0 {
		return nil, &r.dfmError
	}
	return &r.User, nil
}

func parseTime(ut string) string {
	if sec, err := strconv.ParseInt(ut, 10, 64); err == nil {
		t := time.Unix(sec, 0)
		return t.String()
	}
	return ut
}
