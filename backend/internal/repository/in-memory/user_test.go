package in_memory

import (
    "context"
    "testing"

    "github.com/google/uuid"

    "quickflow/internal/models"
    "quickflow/utils"
)

func TestInMemoryUserRepository_SaveUser(t *testing.T) {
    repo := NewInMemoryUserRepository()

    tests := []struct {
        name    string
        user    models.User
        wantErr bool
    }{
        {
            name:    "Save valid user",
            user:    models.User{Id: uuid.New(), Username: "user1", Password: "password", Salt: "salt"},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := repo.SaveUser(context.Background(), tt.user)
            if (err != nil) != tt.wantErr {
                t.Errorf("SaveUser() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestInMemoryUserRepository_GetUser(t *testing.T) {
    repo := NewInMemoryUserRepository()

    salt := utils.GenSalt()
    hashedPassword := utils.HashPassword("password", salt)

    user := models.User{Id: uuid.New(), Username: "user1", Password: hashedPassword, Salt: salt}
    repo.SaveUser(context.Background(), user)

    tests := []struct {
        name       string
        loginData  models.LoginData
        expectedId uuid.UUID
        wantErr    bool
    }{
        {
            name:       "Get existing user",
            loginData:  models.LoginData{Login: "user1", Password: "password"},
            expectedId: user.Id,
            wantErr:    false,
        },
        {
            name:       "Get non-existing user",
            loginData:  models.LoginData{Login: "user2", Password: "password"},
            expectedId: uuid.Nil,
            wantErr:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := repo.GetUser(context.Background(), tt.loginData)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got.Id != tt.expectedId {
                t.Errorf("GetUser() got = %v, want %v", got.Id, tt.expectedId)
            }
        })
    }
}

func TestInMemoryUserRepository_GetUserByUId(t *testing.T) {
    repo := NewInMemoryUserRepository()

    user := models.User{Id: uuid.New(), Username: "user1", Password: "password", Salt: "salt"}
    repo.SaveUser(context.Background(), user)

    tests := []struct {
        name     string
        userId   uuid.UUID
        expected models.User
        wantErr  bool
    }{
        {
            name:     "Get user by valid user ID",
            userId:   user.Id,
            expected: user,
            wantErr:  false,
        },
        {
            name:     "Get user by invalid user ID",
            userId:   uuid.New(),
            expected: models.User{},
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := repo.GetUserByUId(context.Background(), tt.userId)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetUserByUId() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.expected {
                t.Errorf("GetUserByUId() got = %v, want %v", got, tt.expected)
            }
        })
    }
}

func TestInMemoryUserRepository_IsExists(t *testing.T) {
    repo := NewInMemoryUserRepository()

    user := models.User{Id: uuid.New(), Username: "user1", Password: "password", Salt: "salt"}
    repo.SaveUser(context.Background(), user)

    tests := []struct {
        name      string
        login     string
        wantExist bool
    }{
        {
            name:      "User exists",
            login:     "user1",
            wantExist: true,
        },
        {
            name:      "User does not exist",
            login:     "user2",
            wantExist: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            exists := repo.IsExists(context.Background(), tt.login)
            if exists != tt.wantExist {
                t.Errorf("IsExists() got = %v, want %v", exists, tt.wantExist)
            }
        })
    }
}
