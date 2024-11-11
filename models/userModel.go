package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the model that governs all notes objects retrived or inserted into the DB
type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	First_name *string            `json:"first_name"`
	Last_name  *string            `json:"last_name"`
	Address    *string            `json:"address"`
	Password   *string            `json:"password" validate:"required,min=6"`
	Email      *string            `json:"email" validate:"email,required"`
	Phone      string             `json:"phone"`
	Gender     string             `json:"gender" validate:"eq=M|eq=F"`
	PostalCode string             `json:"postal_code"`
	Country    string             `json:"country"`
	Department string             `json:"department"`
	Staff_id   string             `json:"staff_id"`
	Active     bool               `json:"isActive" default:"true"`
	Status     string             `json:"status"`
	// Token         *string            `json:"token"`
	Role        []string `json:"role,omitempty"`
	OTP         string   `json:"otp"`
	OtpVerified bool     `json:"otpVerified" validate:"eq=true|eq=false"`
	// User_type     *string            `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	User_type *string `json:"user_type" validate:"eq=ADMIN|eq=TEACHER|eq=NON_TEACHER|eq=UNBOARDED"`
	// Refresh_token *string            `json:"refresh_token"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	User_id    string    `json:"user_id"`
}

type ChangeUserPassword struct {
	Old_password     *string `json:"old_password" validate:"required"`
	New_password     *string `json:"new_password" validate:"required"`
	Confirm_password *string `json:"confirm_password" validate:"required"`
}

type ChangeStudentPassword struct {
	Email            *string `json:"email" validate:"required"`
	New_password     *string `json:"new_password" validate:"required"`
	Confirm_password *string `json:"confirm_password" validate:"required"`
}

type EditUser struct {
	ID         primitive.ObjectID `bson:"_id"`
	First_name *string            `json:"first_name"`
	Last_name  *string            `json:"last_name"`
	Address    *string            `json:"address"`
	Phone      *string            `json:"phone"`
	PostalCode *string            `json:"postal_code"`
	Country    *string            `json:"country"`
	Department *string            `json:"department"`
	// Token         *string            `json:"token"`
	Role *[]string `json:"role,omitempty"`
	// User_type     *string            `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
	// Refresh_token *string            `json:"refresh_token"`
	Updated_at time.Time `json:"updated_at"`
}

type ValidateUser struct {
	First_name     *string   `json:"first_name" validate:"required"`
	Last_name      *string   `json:"last_name" validate:"required"`
	Phone          *string   `json:"phone" validate:"required"`
	Role           *[]string `json:"role,omitempty" validate:"required"`
	ClassesHandled *[]string `json:"classesHandled"`
	Staff_id       string    `json:"staff_id"`
	Address        *string   `json:"address" validate:"required"`
	User_type      *string   `json:"user_type" validate:"eq=ADMIN|eq=TEACHER|eq=NON_TEACHER"`
	Updated_at     time.Time `json:"updated_at"`
}

type UserReferrer struct {
	ID               primitive.ObjectID `bson:"_id"`
	User_Referrer_Id string             `json:"user_referrer_id"`
	RefereeId        string             `json:"referee_id" validate:"required"`
	Referee_email    string             `json:"referee_email" validate:"email,required"`
	ReferrerId       string             `json:"referrer_id" validate:"required"`
	Referee_type     string             `json:"referee_type" validate:"eq=USER|eq=TEACHER|eq=NON_TEACHER|eq=STUDENT"`
	Referrer_type    string             `json:"referrer_type" validate:"eq=USER|eq=TEACHER|eq=NON_TEACHER|eq=STUDENT"`
	Status           string             `json:"status"`
	Created_at       time.Time          `json:"created_at"`
}

type SignUpUser struct {
	ID         primitive.ObjectID `bson:"_id"`
	User_type  *string            `json:"user_type" validate:"eq=ADMIN|eq=TEACHER|eq=NON_TEACHER|eq=UNBOARDED"`
	Password   *string            `json:"password" validate:"required,min=6"`
	Email      *string            `json:"email" validate:"email,required"`
	User_id    string             `json:"user_id"`
	Status     string             `json:"status"`
	OTP        string             `json:"otp"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
}

type NewUserAlert struct {
	First_name string `json:"first_name" validate:"required"`
	Last_name  string `json:"last_name" validate:"required"`
	Email      string `json:"email" validate:"email,required"`
	User_type  string `json:"user_type" validate:"eq=ADMIN|eq=TEACHER|eq=NON_TEACHER|eq=UNBOARDED"`
}

type SignUpPrecisionUser struct {
	ID          primitive.ObjectID `bson:"_id"`
	Password    *string            `json:"password" validate:"required,min=6"`
	Email       *string            `json:"email" validate:"email,required"`
	User_id     string             `json:"user_id"`
	Is_Verified bool               `json:"is_verified"`
	Status      string             `json:"status"`
	Created_at  time.Time          `json:"created_at"`
	Updated_at  time.Time          `json:"updated_at"`
}

type PrecisionStudentLogin struct {
	Admission_num *string `json:"admission_num" validate:"required"`
}

type Otp struct {
	OTP *string `json:"otp" validate:"required"`
}

type ResendOtp struct {
	Email *string `json:"email" validate:"required"`
}

type OnboardedUserStatus struct {
	ID                  primitive.ObjectID `bson:"_id"`
	User_id             string             `json:"user_id"`
	IsSchoolCompleted   bool               `json:"isSchoolCompleted" default:"false"`
	IsSessionCompleted  bool               `json:"isSessionCompleted" default:"false"`
	IsTeamCompleted     bool               `json:"isTeamCompleted" default:"false"`
	IsSubjectsCompleted bool               `json:"isSubjectsCompleted" default:"false"`
	IsClassCompleted    bool               `json:"isClassCompleted" default:"false"`
	Created_at          time.Time          `json:"created_at"`
	Updated_at          time.Time          `json:"updated_at"`
}
