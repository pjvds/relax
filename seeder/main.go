package main

import(
    "github.com/lancecarlson/couchgo"
    "fmt"
    "os"
    "encoding/json"
    "math/rand"
)

type LookingFor struct {
    Gender string `json:"gender"`
    MinAge int `json:"min-age"`
    MaxAge int `json:"max-age"`
}

type User struct {
    ID string `json:"_id,omitempty"`
    Rev string `json:"_rev,omitempty"`
    Username string `json:"username"`
    Fullname string `json:"fullname"`
    Birthdate string `json:"birthdate"`
    Email string `json:"email"`

    LookingFor *LookingFor `json:"looking-for"`

    // Should always be set to "user" for this type.
    Type string `json:"type"`
}

func main() {
    c, _ := couch.NewClientURL("http://127.0.0.1:5984/relax")
    c.DeleteDB()
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
        user.LookingFor.MinAge = 18 + rand.Intn(99-25) // => 18 && <= 74
        user.LookingFor.MaxAge = 99 - rand.Intn(user.LookingFor.MinAge) // => minAge && <= 99
    }

    fmt.Printf("users loaded\n")
    return
}

func InsertUsers(c *couch.Client, users []*User) (err error) {
    fmt.Printf("insert users into couch\n")

    for i := 0; i < min(100, len(users)); {
        // if _, err = c.Save(users[0]); err != nil {
        //     return
        // }
        chunk := min(10000, len(users)-i)

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

func max(x, y int) int {
    if x > y {
        return x
    }

    return y
}