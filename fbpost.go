
package main
import (
    "net/http"
    "github.com/labstack/echo"
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
    "encoding/json"
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "github.com/dgrijalva/jwt-go"
    "time"
    "github.com/labstack/echo/middleware"
    "./models"
)
func main() {
    e := echo.New()
    
    e.POST("/users", createUser)
    e.GET("/users", getUser)
    e.POST("/login", loginUser)
    r := e.Group("/restricted")
    r.Use(middleware.JWT([]byte("secret")))
    r.POST("/post", createPost)
    r.POST("/comment", createComment)
    r.GET("/post", getPosts)
    e.Logger.Fatal(e.Start(":1323"))
}
type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}
type Token struct{
    Username string
    jwt.StandardClaims
}
func createUser(c echo.Context) error  {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    
    if err != nil {
        panic("failed to connect database")
    }
    db.AutoMigrate(&models.User{})
    
    user_data := &models.User{}
    data_err := json.NewDecoder(c.Request().Body).Decode(&user_data)
    if data_err != nil {
        panic("invalid data")
    }
    
    db.Create(&user_data)
    bytes, err := bcrypt.GenerateFromPassword([]byte(user_data.Password), 14)
    user_data.Password = string(bytes)
    return c.JSON(http.StatusOK, user_data)
    
}
func getUser(c echo.Context) error {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in  connecting database")
    }
    var user models.User
    db.First(&user)
    return c.JSON(http.StatusOK, user)
}

func loginUser(c echo.Context) error {
    var user_data models.User
    data_err := json.NewDecoder(c.Request().Body).Decode(&user_data)
    if data_err != nil {
        var response = map[string]interface{}{"status": false, "message": "Invalid request"}
        return echo.NewHTTPError(http.StatusUnauthorized, response)
    }
    token := _get_token(user_data.Username, user_data.Password)
    var token_str string
    token_str = fmt.Sprint(token)
    return c.JSON(http.StatusOK, map[string]string{
        "token": token_str,
    })
}
func _get_token(username, password string) interface{} {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in  connecting database")
    }
    var user models.User
    if err := db.First(&user, "username = ?", username).Error; err != nil {
        var response = map[string]interface{}{"status": false, "message": "user does not found"}
        return response
    }
    // Create token
    token := jwt.New(jwt.SigningMethodHS256)
    // Set claims
    claims := token.Claims.(jwt.MapClaims)
    claims["id"] = uint(user.ID)
    claims["admin"] = true
    claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
    t, err := token.SignedString([]byte("secret"))
    if err != nil {
        return err
    }
    return t
}
type Post struct {
    Content string `json:"content"`
    PostedAt time.Time `json:"posted_at"`
}
func createPost(c echo.Context) error {
    
    user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    user_id := uint(claims["id"].(float64))
    
    var post_details Post
    data_err := json.NewDecoder(c.Request().Body).Decode(&post_details)
    
    if data_err != nil {
        panic("Invalid input")
    }
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    
    if err != nil {
        panic("error in  connecting database")
    }
    post := &models.Post{Content: post_details.Content, PostedBy: user_id}
    db.AutoMigrate(&models.Post{})
    db.Create(&post)
    var user_data models.User
    db.First(&user_data, user_id)
    db.Model(&user_data).Association("Post").Append(post)
    db.Save(&user_data)
    return c.JSON(http.StatusOK, post)
}
type Comment struct {
    Content string `json:"content"`
    PostId uint `json:"post_id"`
}
func createComment(c echo.Context) error {
    user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    user_id := uint(claims["id"].(float64))
    var comment_details Comment
    data_err := json.NewDecoder(c.Request().Body).Decode(&comment_details)
    if data_err != nil{
        panic("invalid data")
    }
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in  connecting database")
    }
    comment := &models.Comment{Content: comment_details.Content, CommentedBy: user_id, PostID: comment_details.PostId}
    db.AutoMigrate(&models.Comment{})
    db.Create(&comment)
    return c.JSON(http.StatusOK, comment)
}

func getPosts(c echo.Context) error {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in  connecting database")
    }
    var posts []models.Post
    db.Joins("Comment").Find(&posts)
    return c.JSON(http.StatusOK, posts)
}
