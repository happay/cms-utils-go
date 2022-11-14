package util

import (

	"fmt"
	"time"
	"log"
   
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"github.com/asaskevich/govalidator"

	"gorm.io/gorm"
	"os"

	"github.com/dgrijalva/jwt-go"

	"golang.org/x/crypto/bcrypt"
)
type Admin struct {

	gorm.Model
	
	FirstName string `json:"firstName" valid:"required~FirstName is required , length(1|64)~Invalid FirstName"`
	MiddleName  string `json:"middleName" `
	LastName  string `json:"lastName" valid:"required~LastName is required , length(1|64)~Invalid LastName"`
	Email    string `json:"email" valid:"required~Email is required , length(1|20)~Invalid Email"`
	Password string `json:"password" valid:"required~Password is required , length(1|40)~Empty Password"`
	Phone    string `json:"phone" valid:"required~Phone is required , length(1|15)~Invalid Phone"`
	
	
}

func Init() *gorm.DB {
	err := godotenv.Load("jwt.env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	databaseName := os.Getenv("DB_NAME")

	//dbURL := "host= user=postgres password=harshit24; dbname=virtual_account_service port=5432 sslmode=disable"
	dbURL := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, databaseName, port)

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})

	if err != nil {
		panic("Oops! Couldn't connect to the Database.")
	}

	fmt.Printf("success! DB connected! \n")

	err = db.AutoMigrate(&Admin{})
	if err != nil {
		panic(err)
	}




	return db
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
func (m * EmailAlreadyPresentError) Error() string{
    return "User with same email already exists"
}
type PhoneAlreadyPresentError struct {}
func (m *PhoneAlreadyPresentError) Error() string{
	return "User with same phone already exists"
}


func Signup(admin Admin)( error) {
	db := Init()

	
	var admins []Admin
	var Admins []Admin
	

	_, err := govalidator.ValidateStruct(admin)
	if err != nil {
		return err
	}
	
	Encrypted_Password, err := HashPassword(admin.Password)
	if err != nil {

		return err
	}
	admin.Password = Encrypted_Password

	

	db.Where("Email= ?", &admin.Email).Find(&admins)
	
	for i := range admins {
		if admin.Email == admins[i].Email {
			
			
			return &EmailAlreadyPresentError{}
		}
	}

	db.Where("Phone=?", &admin.Phone).Find(&Admins)


	for i := range Admins {
		if admin.Phone == Admins[i].Phone {
			
			return &PhoneAlreadyPresentError{}
		}
	}


	db.Create(&admin)

	

	return nil

}

type WrongPasswordError struct{}

func (m *WrongPasswordError) Error() string {
	return "Invalid Password"
}
func Login(email,password string ) (error ,map[string]string) {
	db := Init()
	result := make(map[string]string)
	var db_Admin Admin
	

	pass1 := password
	

	db.Where("email = ?", email).First(&db_Admin)

	expectedPassword := db_Admin.Password

	if !CheckPasswordHash(pass1, expectedPassword) {
		
		return &WrongPasswordError{},nil
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
		
		return err,nil
	}

	

	result["token"] = tokenString
	result["expires"] = expirationTime.String()

	return nil,result
}


type Claims struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ID       uint   `json:"id"`
	jwt.StandardClaims
}

var jwt_key = []byte("secret_key")