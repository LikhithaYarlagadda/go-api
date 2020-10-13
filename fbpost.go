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
    e.DELETE("/users", deleteUser)
    e.DELETE("/post", deletePost)
    e.PUT("/users", updateUser)
    e.POST("/login", loginUser)
    r := e.Group("/restricted")
    r.Use(middleware.JWT([]byte("secret")))
    e.POST("/post", createPost)
    e.POST("/comment", createComment)
    e.POST("/reply", createReply)
    //e.GET("/post", getPost)
    e.GET("/user/posts", getUserPosts)
    e.POST("/post/react", reactPost)
    e.GET("/reactions", getReactions)
    e.Logger.Fatal(e.Start(":8080"))
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
        panic("error in connecting database")
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
        panic("error in  connecting database")
    }
    var user models.User
    db.First(&user)
    return c.JSON(http.StatusOK, user)
}
func deleteUser(c echo.Context) error {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in connecting database")
    }
    var user models.User
    db.Delete(&user, 2)
    return c.JSON(http.StatusOK, user)
}

type UpdateUser struct{
    Username  string `json:"username"`
}
func updateUser(c echo.Context) error {
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in connecting database")
    }
    var user models.User
    var updated_user UpdateUser
    db.First(&user)
    db.Model(&user).Updates(models.User{Username:updated_user.Username})
    return c.JSON(http.StatusOK, user)
}
func loginUser(c echo.Context) error {
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
        panic("error in connecting database")
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
    UserId uint `json:"user_id"`
}
type PostWithId struct{
    Content string `json:"content"`
    PostedAt time.Time `json:"posted_at"`
    UserId uint `json:"user_id"`
    PostId uint `json:"post_id"`
}
func createPost(c echo.Context) error {
    
    // user := c.Get("user").(*jwt.Token)
    // claims := user.Claims.(jwt.MapClaims)
    // user_id := uint(claims["id"].(float64))
    
    var post_details Post
    data_err := json.NewDecoder(c.Request().Body).Decode(&post_details)
    
    if data_err != nil {
        panic("Invalid input")
    }
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    
    if err != nil {
        panic("error in connecting database")
    }
    post := &models.Post{Content: post_details.Content, PostedBy: post_details.UserId}
    db.AutoMigrate(&models.Post{})
    db.Create(&post)
    var user_data models.User
    db.First(&user_data, post_details.UserId)
    db.Model(&user_data).Association("Post").Append(post)
    db.Save(&user_data)
    return c.JSON(http.StatusOK, post)
}
func deletePost(c echo.Context) error {
    user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    user_id := uint(claims["id"].(float64))
    var post PostWithId
    data_err := json.NewDecoder(c.Request().Body).Decode(&post)
    if data_err != nil{
        panic("Invalid input")
    }
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in connecting database")
    }
    var to_be_deleted_post models.Post
    db.First(&to_be_deleted_post, post.PostId)
    if to_be_deleted_post.PostedBy == user_id{
        panic("Invalid Input")
    }
    db.Delete(&Post{}, post.PostId)
    return c.String(http.StatusOK, "successfully deleted")

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
        panic("Invalid input")
    }
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in connecting database")
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
        panic("error in  connecting database")
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

type User struct {
    UserId uint
    Username string
}

type PostDetails struct {
	PostId        int           `json:"post_id"`
	PostedBy      User          `json:"posted_by"`
	PostContent   string        `json:"post_content"`
	Reactions     Reactions     `json:"reactions"`
	Comments      []CommentDict `json:"comments"`
	CommentsCount int64         `json:"comments_count"`
}

type Reactions struct {
	Count     int64    `json:"count"`
	Reactions []string `json:"reactions"`
}

type CommentDict struct {
	Id             int         `json:"comment_id"`
	CommentedBy    User        `json:"commenter"`
	CommentedAt    string      `json:"commented_at"`
	CommentContent string      `json:"comment_content"`
	ReactionsDict  Reactions   `json:"reactions"`
	Replies        []ReplyDict `json:"replies"`
	RepliesCount   int64       `json:"replies_count"`
}

type ReplyDict struct {
	Id           int64  `json:"comment_id"`
	RepliedBy    User   `json:"commenter"`
	RepliedAt    string `json:"commented_at"`
	ReplyContent string `json:"comment_content"`
}
func getPost(c echo.Context) error {
    var required_post uint
    data_err := json.NewDecoder(c.Request().Body).Decode(&required_post)
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    var post models.Post
    var comments []model.Comment
    var reactions []model.Reaction
    db.First(&post, required_post)
    db.Find(&comments,required_post)
    db.First(&post, required_post)
    var post_dict PostDetails
    post_dict := PostDetails{
        PostId : post.Id,
        Content : post.
    }


    return c.JSON(http.StatusOK, posts)
}
func getUserPosts(c echo.Context) error {
    user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    user_id := uint(claims["id"].(float64))
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("eroor in connecting database")
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
        panic("invalid input")
    }
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in connecting database")
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

func reactComment(c echo.Context) error {
    user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    user_id := uint(claims["id"].(float64))
    var reaction_details Reaction
    data_err := json.NewDecoder(c.Request().Body).Decode(&reaction_details)
    if data_err != nil{
        panic("invalid input")
    }
    db, err := gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
    if err != nil {
        panic("error in connecting database")
    }
    db.AutoMigrate(&models.Reaction{})
    var old_reaction, new_reaction models.Reaction
    error := db.First(&old_reaction, "reacted_by = ? and comment_id = ?", user_id, reaction_details.CommentId).Error
    if error != nil {
        new_reaction = models.Reaction{
            CommentID: reaction_details.CommentId, ReactionType: reaction_details.ReactionType, 
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
        panic("error in connecting database")
    }
    var reactions []ReactionMetrics
    db.Model(&models.Reaction{}).Select("reaction_type, count(reaction_type) as reaction_count").Group("reaction_type").Find(&reactions)
    return c.JSON(http.StatusOK, reactions)
}
