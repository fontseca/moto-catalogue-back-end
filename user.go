package main

type User struct {
  ID          int    `json:"id"`
  FirstName   string `json:"first_name"`
  MiddleName  string `json:"middle_name"`
  LastName    string `json:"last_name"`
  Surname     string `json:"surname"`
  Email       string `json:"email"`
  PhoneNumber string `json:"phone_number"`
  Password    string
  CreatedAt   string `json:"created_at"`
  UpdatedAt   string `json:"updated_at"`
}
