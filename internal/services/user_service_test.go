package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"testing"
	"time"

	"github.com/mlogclub/simple/sqls"
	"golang.org/x/crypto/bcrypt"
)

func seedUserForPasswordTest(t *testing.T, id int64, username, password string) {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate password hash failed: %v", err)
	}

	now := time.Now()
	user := &models.User{
		ID:           id,
		Username:     username,
		Nickname:     username,
		Password:     string(passwordHash),
		Status:       1,
		PasswordSalt: "",
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   id,
			CreateUserName: username,
			UpdatedAt:      now,
			UpdateUserID:   id,
			UpdateUserName: username,
		},
	}
	if err := sqls.DB().Create(user).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
}

func TestChangeOwnPassword(t *testing.T) {
	setupIMServiceTestDB(t)
	seedUserForPasswordTest(t, 3001, "self_user", "old_password")

	operator := &dto.AuthPrincipal{
		UserID:   3001,
		Username: "self_user",
	}
	if err := UserService.ChangeOwnPassword("new_password", operator); err != nil {
		t.Fatalf("change own password failed: %v", err)
	}

	updated := UserService.Get(3001)
	if updated == nil {
		t.Fatal("expected updated user")
	}
	if bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte("new_password")) != nil {
		t.Fatal("expected password to be updated")
	}
}

func TestChangeOwnPasswordRequiresLogin(t *testing.T) {
	setupIMServiceTestDB(t)

	if err := UserService.ChangeOwnPassword("new_password", nil); err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestChangeOwnPasswordRejectsEmptyPassword(t *testing.T) {
	setupIMServiceTestDB(t)
	seedUserForPasswordTest(t, 3002, "blank_user", "old_password")

	operator := &dto.AuthPrincipal{
		UserID:   3002,
		Username: "blank_user",
	}
	if err := UserService.ChangeOwnPassword("   ", operator); err == nil {
		t.Fatal("expected invalid password error")
	}
}

func TestResetPasswordGeneratesRandomPassword(t *testing.T) {
	setupIMServiceTestDB(t)
	seedUserForPasswordTest(t, 3003, "reset_user", "old_password")

	operator := &dto.AuthPrincipal{
		UserID:   1,
		Username: "admin",
	}
	password, err := UserService.ResetPassword(3003, operator)
	if err != nil {
		t.Fatalf("reset password failed: %v", err)
	}
	if len(password) != 12 {
		t.Fatalf("expected 12-char password, got %d", len(password))
	}

	updated := UserService.Get(3003)
	if updated == nil {
		t.Fatal("expected updated user")
	}
	if bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte(password)) != nil {
		t.Fatal("expected password to be updated with generated password")
	}
}
