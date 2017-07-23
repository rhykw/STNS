package stns

import (
	"net/http"
	"strconv"

	"github.com/ant0ine/go-json-rest/rest"
)

type responser interface {
	Response(*Config)
}

// ----------------------------------------
// v1
// ----------------------------------------
type v1ResponseFormat struct {
	Items Attributes
	query *Query
	w     rest.ResponseWriter
	r     *rest.Request
}

func (res *v1ResponseFormat) Response(c *Config) {
	if res.Items == nil {
		rest.NotFound(res.w, res.r)
	} else {
		res.w.WriteJson(res.Items)
	}
}

// ----------------------------------------
// v2
// ----------------------------------------
type v2MetaData struct {
	APIVersion float64 `json:"api_version"`
	Result     string  `json:"result"`
}

type v2ResponseFormat struct {
	MetaData *v2MetaData `json:"metadata"`
	Items    Attributes  `json:"items"`
	query    *Query
	w        rest.ResponseWriter
	r        *rest.Request
}

func (res *v2ResponseFormat) Response(c *Config) {
	if res.Items == nil {
		res.w.WriteHeader(http.StatusNotFound)
	}

	response := v2ResponseFormat{
		MetaData: &v2MetaData{
			APIVersion: 2.1,
			Result:     "success",
		},
		Items: res.Items,
	}
	res.w.WriteJson(response)
	return
}

// ----------------------------------------
// v3
// ----------------------------------------
type v3ResponseFormat struct {
	Items Attributes
	query *Query
	w     rest.ResponseWriter
	r     *rest.Request
}

type v3User struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Password      string   `json:"password"`
	GroupID       int      `json:"group_id"`
	Directory     string   `json:"directory"`
	Shell         string   `json:"shell"`
	Gecos         string   `json:"gecos"`
	Keys          []string `json:"keys"`
	SetupCommands []string `json:"setup_commands"`
}

type v3Group struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Users []string `json:"users"`
}

type v3Sudo struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type v3Users struct {
	items []*v3User
}
type v3Groups struct {
	items []*v3Group
}

type v3Sudoers struct {
	items []*v3Sudo
}

type v3Resource interface {
	buildResource(string, *Attribute) interface{}
}

func newV3Resource(q *Query) v3Resource {
	switch q.resource {
	case "user":
		return v3Users{}
	case "group":
		return v3Groups{}
	case "sudo":
		return v3Sudoers{}
	}
	return nil
}

func setPrevNextHeader(res *v3ResponseFormat, c *Config) {
	if res.query.column == "id" {
		switch res.query.resource {
		case "user":
			res.w.Header().Set("STNS-PREV-ID", strconv.Itoa(c.UserMinID))
			res.w.Header().Set("STNS-NEXT-ID", strconv.Itoa(c.UserMaxID))
		case "group":
			res.w.Header().Set("STNS-PREV-ID", strconv.Itoa(c.GroupMinID))
			res.w.Header().Set("STNS-NEXT-ID", strconv.Itoa(c.GroupMaxID))
		}
	}

}

func (user v3Users) buildResource(n string, u *Attribute) interface{} {
	if n != "" && u.ID != 0 {
		user := &v3User{
			Name: n,
			ID:   u.ID,
		}

		if u.User != nil {
			user.Password = u.Password
			user.GroupID = u.GroupID
			user.Directory = u.Directory
			user.Shell = u.Shell
			user.Gecos = u.Gecos
			user.Keys = u.Keys
			user.SetupCommands = u.SetupCommands
		}
		return user
	}
	return nil
}

func (user v3Groups) buildResource(n string, g *Attribute) interface{} {
	if g.ID != 0 {
		group := &v3Group{
			Name: n,
			ID:   g.ID,
		}

		if g.Group != nil {
			group.Users = g.Users
		}
		return group
	}
	return nil
}

func (user v3Sudoers) buildResource(n string, u *Attribute) interface{} {
	if u.User != nil {
		return &v3Sudo{
			Name:     n,
			Password: u.Password,
		}
	}
	return nil
}

func (res *v3ResponseFormat) Response(c *Config) {
	if len(res.Items) == 0 {
		setPrevNextHeader(res, c)
		rest.NotFound(res.w, res.r)
		return
	}

	resource := newV3Resource(res.query)
	resources := []interface{}{}

	for n, u := range res.Items {
		r := resource.buildResource(n, u)

		if r != nil {
			resources = append(resources, r)
		}

		if res.query.column != "list" {
			break
		}
	}

	if len(resources) > 0 {
		if res.query.column == "list" {
			res.w.WriteJson(resources)
		} else {
			res.w.WriteJson(resources[0])
		}
	} else {
		rest.NotFound(res.w, res.r)
	}
}

func newResponder(q *Query, w rest.ResponseWriter, r *rest.Request) responser {
	res := q.Get()
	switch r.URL.Path[1:3] {
	case "v2":
		return &v2ResponseFormat{
			Items: res,
			query: q,
			w:     w,
			r:     r,
		}
	case "v3":
		return &v3ResponseFormat{
			Items: res,
			query: q,
			w:     w,
			r:     r,
		}
	default:
		return &v1ResponseFormat{
			Items: res,
			query: q,
			w:     w,
			r:     r,
		}
	}
}
