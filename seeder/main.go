package main

import(
    "github.com/lancecarlson/couchgo"
    "fmt"
    "os"
    "encoding/json"
    "time"
)

type User struct {
    ID string `json:"_id,omitempty"`
    Rev string `json:"_rev,omitempty"`
    Username string `json:"username"`
    Fullname string `json:"fullname"`
    BirthDate time.Time `json:"birthdate"`
    Email string `json:"email"`

    DocType string `json:"doc-type"`
}

func main() {
    c, _ := couch.NewClientURL("http://127.0.0.1:5984/relax")
    c.CreateDB()

    users, err := GenerateUsers()
    if err != nil {
        fmt.Printf("cannot generate users: %v\n", err)
        return
    }
    fmt.Printf("generate %v users\n", len(users))

    if err = InsertUsers(c, users); err != nil {
        fmt.Printf("error inserting users: %v\n", err)
        return
    }
}

func GenerateUsers() (users []*User, err error) {
    var file *os.File
    if file, err = os.Open("users.json"); err != nil {
        return
    }

    fmt.Printf("loading user data into memory\n")
    decoder := json.NewDecoder(file)
    if err = decoder.Decode(&users); err != nil {
        return
    }

    for _, user := range users {
        user.DocType = "user"
    }

    fmt.Printf("users loaded\n")
    return
}

func InsertUsers(c *couch.Client, users []*User) (err error) {
    fmt.Printf("insert users into couch\n")

    for i := 0; i < len(users); {
        // if _, err = c.Save(users[0]); err != nil {
        //     return
        // }
        chunk := min(1000, len(users)-i)

        var bulk []interface{}
        for _, user := range users[i:i+chunk] {
            bulk = append(bulk, user)
        }

        _, _, err = c.BulkSave(bulk...)
        if err != nil {
            return
        }
        i += chunk

        if i % 10 == 0 {
            fmt.Printf("%v users saved\n", i)
        }
    }

    fmt.Printf("%v users saved\n", len(users))
    fmt.Printf("all users saved\n")
    return
}

func min(x, y int) int {
    if x < y {
        return x
    }

    return y
}