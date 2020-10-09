
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
    //   "reflect"
    //   "errors"
)
func main() {
    e := echo.New()
    
    e.POST("/users", createUser)
    e.GET("/users", getUser)
    e.DELETE("/users", deleteUser)
    e.PUT("/users", updateUser)
    e.POST("/login", login)
    r := e.Group("/restricted")
    r.Use(middleware.JWT([]byte("secret")))
    r.POST("/post", createPost)
    r.POST("/comment", createComment)
    r.POST("/reply", createReply)
    r.GET("/post", getPosts)
    r.GET("/user/posts", getUserPosts)
    r.POST("/post/react", reactPost)
    r.GET("/reactions", getReactions)
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
        panic("wrong data")
    }
    
    db.Create(&user_data)
    bytes, err := bcrypt.GenerateFromPassword([]byte(user_data.Password), 14)
    user_data.Password = string(bytes)
    return c.JSON(http.StatusOK, user_data)
    
}
func getUser(c echo.Context) error {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    var user models.User
    db.First(&user)
    return c.JSON(http.StatusOK, user)
}
func deleteUser(c echo.Context) error {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    var user models.User
    db.Delete(&user, 2)
    return c.JSON(http.StatusOK, user)
}
func updateUser(c echo.Context) error {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    var user models.User
    db.First(&user)
    db.Model(&user).Updates(models.User{Username: "update"})
    return c.JSON(http.StatusOK, user)
}
func login(c echo.Context) error {
    var user_data models.User
    data_err := json.NewDecoder(c.Request().Body).Decode(&user_data)
    if data_err != nil {
        var resp = map[string]interface{}{"status": false, "message": "Invalid request"}
        return echo.NewHTTPError(http.StatusUnauthorized, resp)
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
        panic("failed to connect database")
    }
    var user models.User
    if err := db.First(&user, "username = ?", username).Error; err != nil {
        var resp = map[string]interface{}{"status": false, "message": "user does not found"}
        return resp
    }
    // Create token
    token := jwt.New(jwt.SigningMethodHS256)
    // Set claims
    claims := token.Claims.(jwt.MapClaims)
    claims["id"] = uint(user.ID)
    claims["admin"] = true
    claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
    // Generate encoded token and send it as response.
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
        panic("failed to connect database")
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
        panic("failed to connect database")
    }
    comment := &models.Comment{Content: comment_details.Content, CommentedBy: user_id, PostID: comment_details.PostId}
    db.AutoMigrate(&models.Comment{})
    db.Create(&comment)
    return c.JSON(http.StatusOK, comment)
}
type Reply struct {
    Content string `json:"content"`
    CommentId uint `json:"comment_id"`
}
func createReply(c echo.Context) error {
    user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    user_id := uint(claims["id"].(float64))
    var reply_details Reply
    data_err := json.NewDecoder(c.Request().Body).Decode(&reply_details)
    if data_err != nil{
        panic("invalid data")
    }
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    var comment models.Comment
    db.Find(&comment, reply_details.CommentId)
    reply := &models.Comment{
        Content: reply_details.Content, PostID: comment.PostID,
        CommentedBy: user_id,
    }
    
    db.Create(&reply)
    db.Model(&comment).Association("Replies").Append(reply)
    db.Save(&comment)
    return c.JSON(http.StatusOK, reply)
}
// type Comments struct {
//     CommentID uint
//     Content string
//     CommentedBy uint
// }
// type Posts struct {
//     PostID uint
//     Content string
//     PostedBy uint
//     Comment []Comments
// }
func getPosts(c echo.Context) error {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    var posts []models.Post
    // db.Table("posts").Joins("left join comments on comments.post_id = posts.id").Scan(&posts)
    db.Joins("Comment").Find(&posts)
    return c.JSON(http.StatusOK, posts)
}
func getUserPosts(c echo.Context) error {
    user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    user_id := uint(claims["id"].(float64))
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    var posts []models.Post
    db.Where("posted_by = ?", user_id).Find(&posts)
    return c.JSON(http.StatusOK, posts)
}
type Reaction struct {
    PostId uint `json:"post_id"`
    CommentId uint `json:"comment_id"`
    ReactionType string `json:"reaction_type"`
}
func reactPost(c echo.Context) error {
    user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    user_id := uint(claims["id"].(float64))
    var reaction_details Reaction
    data_err := json.NewDecoder(c.Request().Body).Decode(&reaction_details)
    if data_err != nil{
        panic("invalid data")
    }
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    db.AutoMigrate(&models.Reaction{})
    var old_reaction, new_reaction models.Reaction
    error := db.First(&old_reaction, "reacted_by = ? and post_id = ?", user_id, reaction_details.PostId).Error
    if error != nil {
        new_reaction = models.Reaction{
            PostID: reaction_details.PostId, ReactionType: reaction_details.ReactionType, 
            ReactedBy: user_id,
        }
        db.Create(&new_reaction)
        return c.String(http.StatusOK, "successfully created")
    }
    is_delete := old_reaction.ReactionType == reaction_details.ReactionType
    if is_delete{
        db.Delete(&old_reaction, old_reaction.ID)
    } else {
        old_reaction.ReactionType = reaction_details.ReactionType
        db.Save(&old_reaction)
    }
    return c.String(http.StatusOK, "successfully created")
    
}
type ReactionMetrics struct{
    ReactionType string `json:"reaction_type"`
    ReactionCount int `json:"reaction_count"`
}
func getReactions(c echo.Context) error {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    var reactions []ReactionMetrics
    // reactions :=[]type map[string]interface{}
    db.Model(&models.Reaction{}).Select("reaction_type, count(reaction_type) as reaction_count").Group("reaction_type").Find(&reactions)
    return c.JSON(http.StatusOK, reactions)
}
