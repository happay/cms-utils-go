package util

import (
	"time"

	"github.com/asaskevich/govalidator"

	"gorm.io/gorm"

	"github.com/dgrijalva/jwt-go"

	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	gorm.Model

	FirstName  string `json:"firstName" valid:"required~FirstName is required "`
	MiddleName string `json:"middleName" `
	LastName   string `json:"lastName" valid:"required~LastName is required "`
	Email      string `json:"email" valid:"required~Email is required "`
	Password   string `json:"password" valid:"required~Password is required "`
	Phone      string `json:"phone" valid:"required~Phone is required "`
}

func AutoMigrate(db *gorm.DB) {

	err := db.AutoMigrate(&Admin{})
	if err != nil {
		panic(err)
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return emailRegex.MatchString(e)
}

type InvalidEmailError struct{}

func (m *InvalidEmailError) Error() string {
	return "Invalid Email"
}

type InvalidPasswordError struct{}

func (m *InvalidPasswordError) Error() string {
	return "Invalid Password"
}

type InvalidPhoneError struct{}

func (m *InvalidPhoneError) Error() string {
	return "Invalid Phone number"
}

type InvalidFirstnameError struct{}

func (m *InvalidFirstnameError) Error() string {
	return "Invalid Firstname"
}

type InvalidLastnameError struct{}

func (m *InvalidLastnameError) Error() string {
	return "Invalid Lastname"
}

type InvalidMiddlenameError struct{}

func (m *InvalidMiddlenameError) Error() string {
	return "Invalid Middlename"
}
func isPhoneValid(s string) bool {
	phoneRegex := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){10,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
	return phoneRegex.MatchString(s)
}
func verifyPassword(s string) bool {
	letters := 0
	var sixOrMore, number, upper, special, lower bool
	for _, c := range s {
		letters++
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true

		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true

		case unicode.IsLower(c):
			lower = true

		default:
			//return false, false, false, false
		}
	}

	sixOrMore = letters >= 6
	return sixOrMore && number && upper && special && lower
}
func isNameValid(s string) bool {
	nameRegex := regexp.MustCompile("^[A-Za-z][A-Za-z0-9_]{0,}$")
	return nameRegex.MatchString(s)
}

func Signup(admin Admin, db *gorm.DB,expiry_time int) (error,Admin,error,map[string]string) {

	_, err := govalidator.ValidateStruct(admin)
	if err != nil {
		return err, admin,nil,nil
	}

	if !isEmailValid(admin.Email) {
		return &InvalidEmailError{}, admin,nil,nil
	}
	if !isPhoneValid(admin.Phone) {
		return &InvalidPhoneError{}, admin,nil,nil
	}

	if !(verifyPassword(admin.Password)) {
		return &InvalidPasswordError{}, admin,nil,nil
	}
	if !isNameValid(admin.FirstName) {
		return &InvalidFirstnameError{}, admin,nil,nil
	}
	if !isNameValid(admin.LastName) {
		return &InvalidLastnameError{}, admin,nil,nil
	}
	if admin.MiddleName != "" {
		if !isNameValid(admin.MiddleName) {
			return &InvalidMiddlenameError{}, admin,nil,nil
		}
	}
    err,token:=Login(admin.Email,admin.Password,db,expiry_time)
	if err!=nil{
		return nil,admin,err,nil
	}
	return nil, admin,nil,token

}

type WrongPasswordError struct{}

func (m *WrongPasswordError) Error() string {
	return "Wrong Password"
}
func Login(email, password string, db *gorm.DB, expiry_time int) (error, map[string]string) {

	result := make(map[string]string)
	var db_Admin Admin

	pass1 := password

	db.Where("email = ?", email).First(&db_Admin)

	expectedPassword := db_Admin.Password

	if !CheckPasswordHash(pass1, expectedPassword) {

		return &WrongPasswordError{}, nil
	}

	expirationTime := time.Now().Add(time.Minute * time.Duration(expiry_time))
	claimsForFiniteTime := &Claims{
		Email:    email,
		Password: password,
		ID:       db_Admin.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	claimsForInfiniteTime := &Claims{
		Email:    email,
		Password: password,
		ID:       db_Admin.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: 0,
		},
	}
	var token *jwt.Token
	if expiry_time != 0 {
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claimsForFiniteTime)
	} else {
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claimsForInfiniteTime)
	}
	tokenString, err := token.SignedString(jwt_key)
	if err != nil {

		return err, nil
	}

	result["token"] = tokenString
	if expiry_time != 0 {
		result["expires"] = expirationTime.String()
	} else {
		result["expires"] = "Infinite"
	}

	return nil, result
}

type Claims struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ID       uint   `json:"id"`
	jwt.StandardClaims
}
type TokenNilError struct{}

func (m *TokenNilError) Error() string {
	return "Token is nil."
}

type TokenInvalidError struct{}

func (m *TokenInvalidError) Error() string {
	return "Token is Invalid."
}

var jwt_key = []byte("secret_key")

func TokenValidation(tokenStr string) (bool, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) { return jwt_key, nil })

	if tkn == nil {

		return false, &TokenNilError{}
	}

	if err != nil {
		if err == jwt.ErrSignatureInvalid {

			return false, err
		} else {

			return false, err
		}
	}

	if !tkn.Valid {

		return false, &TokenInvalidError{}
	}
	return true, nil
}
func Refresh_token(tokenStr string, expiry_time int) (error, map[string]string) {

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return jwt_key, nil
		})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {

			return err, nil
		}

		return err, nil
	}
	if !tkn.Valid {

		return &TokenInvalidError{}, nil
	}

	expirationTime := time.Now().Add(time.Minute * time.Duration(expiry_time))
	if expiry_time == 0 {
		claims.ExpiresAt = 0
	} else {
		claims.ExpiresAt = expirationTime.Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwt_key)

	if err != nil {

		return err, nil
	}

	result := make(map[string]string)

	result["refresh_token"] = tokenString
	if expiry_time != 0 {
		result["expires"] = expirationTime.String()
	} else {
		result["expires"] = "Infinite"
	}
	return nil,result
}
