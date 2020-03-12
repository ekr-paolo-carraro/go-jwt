package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"

	"github.com/ekr-paolo-carraro/go-jwt/domain"
	"github.com/ekr-paolo-carraro/go-jwt/service"
	"github.com/gin-gonic/gin"
)

type HandlerRouter struct {
	Service service.PostgresService
}

func NewRouter() (*gin.Engine, error) {

	service, err := service.InitDBService()
	if err != nil {
		return nil, fmt.Errorf("error on init db service: %v", err)
	}
	hr := HandlerRouter{*service}

	router := gin.Default()
	router.POST("/signup", hr.signupHandler)
	router.POST("/login", hr.loginHandler)
	router.GET("/protected", hr.authMiddleware(), hr.protectedEndpointHandler)

	return router, nil
}

func (hr HandlerRouter) sendErrorMessage(c *gin.Context, status int, message string) {
	var err domain.Error
	err.Message = message

	c.JSON(status, err)
}

func (hr HandlerRouter) parseUser(c *gin.Context) (*domain.User, error) {
	var user domain.User
	if err := c.BindJSON(&user); err != nil {
		return nil, err
	}

	if user.Email == "" {
		return nil, errors.New("Email can't be empty")
	}

	re := regexp.MustCompile(`^[A-Za-z0-9._\-\+]+@[A-Za-z0-9._\-\+]+\..{2,}$`)
	if ok := re.Match([]byte(user.Email)); ok == false {
		return nil, errors.New("Email is incorrect")
	}

	if len(user.Password) < 8 {
		return nil, errors.New("Password it's too short (min 8 char)")
	}

	return &user, nil
}

func (hr HandlerRouter) signupHandler(c *gin.Context) {

	user, err := hr.parseUser(c)
	if err != nil {
		hr.sendErrorMessage(c, http.StatusBadRequest, "Error on user data: "+err.Error())
		return
	}

	candidate, err := hr.Service.GetUser(user.Email)
	if err != nil {
		hr.sendErrorMessage(c, http.StatusBadRequest, "Error on check user existence: "+err.Error())
		return
	}

	if candidate != nil {
		hr.sendErrorMessage(c, http.StatusPreconditionFailed, "User already exists: "+user.Email)
		return
	}

	newid, err := hr.Service.AddUser(*user)
	if err != nil {
		hr.sendErrorMessage(c, http.StatusBadRequest, "Error on insert user: "+err.Error())
		return
	}

	user.ID = newid
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

func (hr HandlerRouter) generateToken(user domain.User) (string, error) {
	var err error
	var key string = os.Getenv("KEY")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"iss":   "test",
	})

	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (hr HandlerRouter) loginHandler(c *gin.Context) {

	user, err := hr.parseUser(c)
	if err != nil {
		hr.sendErrorMessage(c, http.StatusBadRequest, "Error on user data: "+err.Error())
		return
	}

	tk, err := hr.generateToken(*user)
	if err != nil {
		hr.sendErrorMessage(c, http.StatusInternalServerError, err.Error())
	}

	dbUser, err := hr.Service.GetUser(user.Email)
	if err != nil {
		hr.sendErrorMessage(c, http.StatusInternalServerError, err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
	if err != nil {
		hr.sendErrorMessage(c, http.StatusUnauthorized, "User not authrized")
		return
	}

	c.JSON(http.StatusOK, domain.JWT{
		Token: tk,
	})
}

func (hr HandlerRouter) protectedEndpointHandler(c *gin.Context) {
	c.String(http.StatusOK, "hello protectedEndpointHandler")
}

func (hr HandlerRouter) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		var token string
		if strings.Contains(authHeader, " ") {
			token = strings.Split(authHeader, " ")[1]
		} else {
			hr.sendErrorMessage(c, http.StatusUnauthorized, "Error: auth not valid")
			c.Abort()
			return
		}

		parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Error on parse token")
			}
			return []byte(os.Getenv("KEY")), nil
		})

		if err != nil {
			hr.sendErrorMessage(c, http.StatusUnauthorized, "Error: auth not valid")
			c.Abort()
			return
		}

		if !parsedToken.Valid {
			hr.sendErrorMessage(c, http.StatusUnauthorized, "Error: auth not valid")
			c.Abort()
			return
		}
		c.Next()
	}
}
