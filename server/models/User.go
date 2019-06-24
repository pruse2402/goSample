package models

import (
	"regexp"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"gosample/server/validation"
)

type UserRole string

const (
	Admin  = "Admin"
	Normal = "Normal"
)

type User struct {
	ID           bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name         string        `json:"name" bson:"name"`
	Role         UserRole      `json:"role" bson:"role"`
	Email        string        `json:"email" bson:"email"`
	Age          int           `json:"age" bson:"age"`
	Password     string        `json:"password" bson:"password"`
	DOB          time.Time     `json:"-" bson:"dob"`
	DOBStr       string        `json:"dob" bson:"-"`
	Image        []byte        `json:"-" bson:"image"`
	DateCreated  time.Time     `json:"dateCreated" bson:"dateCreated"`
	LastUpdated  time.Time     `json:"lastUpdated" bson:"lastUpdated"`
	ActiveStatus bool          `json:"activeStatus" bson:"activeStatus"`
}

func (user *User) Validate(db *mgo.Database) (bool, map[string]interface{}) {

	v := &validation.Validation{}

	res := v.Required(user.Name).Key("name").Message("Enter name")
	if res.Ok {
		v.MaxSize(user.Name, 50).Key("name").Key("Name should not be more than 5 characters")
	}

	res = v.Required(user.Email).Key("email").Message("Enter email")
	if res.Ok {
		res = v.Email(user.Email).Key("email").Message("Enter valid email")
		if res.Ok {
			query := bson.M{"email": bson.RegEx{"^" + user.Email + "$", "i"}}
			if user.ID.Hex() != "" {
				query["_id"] = bson.M{"$ne": user.ID}
			}
			c, _ := db.C("user").Find(query).Count()
			if c != 0 {
				v.Error("Email already exists").Key("email")
			}
		}
	}

	res = v.Required(user.Role).Key("role").Message("Enter role")
	if res.Ok {
		if !(user.Role == Admin || user.Role == Normal) {
			v.Error("Enter valid role").Key("role")
		}
	}

	res = v.Required(user.Age).Key("age").Message("Enter age")
	if res.Ok {
		v.Min(user.Age, 18).Key("age").Message("Age should be from 18")
		v.Max(user.Age, 100).Key("age").Message("Age should not be greater than 100")
	}

	res = v.Required(user.Password).Key("password").Message("Enter password")
	if res.Ok {
		v.Match(user.Password, regexp.MustCompile("^[0-9A-Za-z]+$")).Key("password").Message("Password should be alphanumeric")
	}

	return v.HasErrors(), v.ErrorMap()
}
