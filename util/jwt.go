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
func verifyPassword(s string) (sixOrMore, number, upper, special,lower bool) {
    letters := 0
	
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
			lower=true

        default:
            //return false, false, false, false
        }
    }
	
    sixOrMore = letters >= 6
    return
}
func isNameValid(s string) bool {
	nameRegex := regexp.MustCompile("^[A-Za-z][A-Za-z0-9_]{0,}$")
	return nameRegex.MatchString(s)
}

func Signup(admin Admin) (error, Admin) {

	_, err := govalidator.ValidateStruct(admin)
	if err != nil {
		return err, admin
	}
 
	if !isEmailValid(admin.Email) {
		return &InvalidEmailError{}, admin
	}
	if !isPhoneValid(admin.Phone) {
		return &InvalidPhoneError{}, admin
	}
	sixOrMore,number,upper,special,lower:=verifyPassword(admin.Password)

	if !(sixOrMore && number && upper && special && lower) {
		return &InvalidPasswordError{}, admin
	}
	if !isNameValid(admin.FirstName) {
		return &InvalidFirstnameError{}, admin
	}
	if !isNameValid(admin.LastName) {
		return &InvalidLastnameError{}, admin
	}
	if admin.MiddleName != "" {
		if !isNameValid(admin.MiddleName) {
			return &InvalidMiddlenameError{}, admin
		}
	}

	return nil, admin

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
	claims := &Claims{
		Email:    email,
		Password: password,
		ID:       db_Admin.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwt_key)
	if err != nil {

		return err, nil
	}

	result["token"] = tokenString
	result["expires"] = expirationTime.String()

	return nil, result
}

type Claims struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ID       uint   `json:"id"`
	jwt.StandardClaims
}

var jwt_key = []byte("secret_key")
