package util

import (
	"time"

	"github.com/asaskevich/govalidator"

	"gorm.io/gorm"

	"github.com/dgrijalva/jwt-go"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	gorm.Model

	FirstName  string `json:"firstName" valid:"required~FirstName is required , length(1|64)~Invalid FirstName"`
	MiddleName string `json:"middleName" `
	LastName   string `json:"lastName" valid:"required~LastName is required , length(1|64)~Invalid LastName"`
	Email      string `json:"email" valid:"required~Email is required , length(1|20)~Invalid Email"`
	Password   string `json:"password" valid:"required~Password is required , length(1|40)~Empty Password"`
	Phone      string `json:"phone" valid:"required~Phone is required , length(1|15)~Invalid Phone"`
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

type EmailAlreadyPresentError struct {
}

func (m *EmailAlreadyPresentError) Error() string {
	return "User with same email already exists"
}

type PhoneAlreadyPresentError struct{}

func (m *PhoneAlreadyPresentError) Error() string {
	return "User with same phone already exists"
}

func Signup(admin Admin, db *gorm.DB) (error,*gorm.DB)  {

	var admins []Admin
	var Admins []Admin

	_, err := govalidator.ValidateStruct(admin)
	if err != nil {
		return err,nil
	}

	Encrypted_Password, err := HashPassword(admin.Password)
	if err != nil {

		return err,nil
	}
	admin.Password = Encrypted_Password

	db.Where("Email= ?", &admin.Email).Find(&admins)

	for i := range admins {
		if admin.Email == admins[i].Email {

			return &EmailAlreadyPresentError{},nil
		}
	}

	db.Where("Phone=?", &admin.Phone).Find(&Admins)

	for i := range Admins {
		if admin.Phone == Admins[i].Phone {

			return &PhoneAlreadyPresentError{},nil
		}
	}

	db.Create(&admin)

	return nil,admin

}

type WrongPasswordError struct{}

func (m *WrongPasswordError) Error() string {
	return "Invalid Password"
}
func Login(email, password string, db *gorm.DB) (error, map[string]string) {

	result := make(map[string]string)
	var db_Admin Admin

	pass1 := password

	db.Where("email = ?", email).First(&db_Admin)

	expectedPassword := db_Admin.Password

	if !CheckPasswordHash(pass1, expectedPassword) {

		return &WrongPasswordError{}, nil
	}

	expirationTime := time.Now().Add(time.Minute * 30)
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
