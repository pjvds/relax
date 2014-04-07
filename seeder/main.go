package main

import(
    "github.com/lancecarlson/couchgo"
    "fmt"
    "os"
    "encoding/json"
    "math/rand"
    "time"
)

type LookingFor struct {
    Gender string `json:"gender"`
    MinAge int `json:"min-age"`
    MaxAge int `json:"max-age"`
}

type ProfileUpdate struct {
    ID string `json:"_id,omitempty"`
    Rev string `json:"_rev,omitempty"`

    User string `json:"user"`
    Datetime string `json:"datetime"`
    Text string `json:"text"`

    // Should always be set to "profile-update" for this type.
    Collection string `json:"collection"`        
}

type User struct {
    ID string `json:"_id,omitempty"`
    Rev string `json:"_rev,omitempty"`
    Username string `json:"username"`
    Fullname string `json:"fullname"`
    Gender string `json:"gender"`
    BirthDate string `json:"birthdate"`
    BirthYear int `json:"birthyear"`
    Email string `json:"email"`

    LookingFor *LookingFor `json:"looking-for"`

    // Should always be set to "user" for this type.
    Collection string `json:"collection"`
}

func main() {
    c, _ := couch.NewClientURL("http://127.0.0.1:5984/relax")
    c.DeleteDB()
    c.CreateDB()

    InsertUsers(c)
    InsertProfileUpdates(c)
}

func InsertProfileUpdates(c *couch.Client) (err error) {
    var updates []*ProfileUpdate
    var file *os.File
    
    if file, err = os.Open("user-updates.json"); err != nil {
        fmt.Printf("error opening user-updates.json: %v\n", err)
        return
    }

    decoder := json.NewDecoder(file)
    if err = decoder.Decode(&updates); err != nil {
        fmt.Printf("error decoding user updates: %v\n", err)
        return
    }

    documents := make([]interface{}, len(updates), len(updates))
    for k, v := range updates {
        documents[k] = v
    }
    saveDocuments(c, documents)
    return
}

func InsertUsers(c *couch.Client) (err error) {
    var users []*User

    if users, err = generateUsers(); err != nil {
        fmt.Printf("cannot generate users: %v\n", err)
        return
    }

    documents := make([]interface{}, len(users), len(users))
    for k, v := range users {
        documents[k] = v
    }

    if err = saveDocuments(c, documents); err != nil {
        fmt.Printf("error inserting users: %v\n", err)
        return
    }
    fmt.Printf("generate %v users\n", len(users))
    return
}

func ExportUserIds(users []*User) {
    filename := "user-ids.txt"
    file, err := os.OpenFile(filename, os.O_CREATE | os.O_TRUNC | os.O_WRONLY, os.ModePerm)
    if err != nil {
        fmt.Printf("cannot create user-ids.txt: %v\n", err)
    }
    defer file.Close()

    for _, user := range users {
        file.Write([]byte(user.ID + "\n"))
    }

    if err := file.Sync(); err != nil {
        fmt.Printf("cannot create user-ids.txt: %v\n", err)
    }

    fmt.Printf("written user ids to user-ids.txt\n")
}

func generateUsers() (users []*User, err error) {
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
    
        var date time.Time
        date, err = time.Parse("2006-01-15", user.BirthDate)
        user.BirthYear = date.Year()
    }

    fmt.Printf("users loaded\n")
    return
}

func saveDocuments(c *couch.Client, documents []interface{}) (err error) {
    fmt.Printf("insert documents into couch\n")

    for i := 0; i < len(documents); {
        chunk := min(15*1000, len(documents)-i)
        bulk := documents[i:i+chunk]

        _, _, err = c.BulkSave(bulk...)
        if err != nil {
            return
        }
        i += chunk

        if i % 10 == 0 {
            fmt.Printf("%v documents saved\n", i)
        }
    }

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
