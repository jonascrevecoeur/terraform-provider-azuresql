package sql

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

var (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	SpecialCharSet = "!@#$%&*"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + SpecialCharSet + numberSet
)

func generatePassword(passwordLength int, minSpecialChars int, minNum int, minUpperCase int, allowedSpecialChars string) string {
	var password strings.Builder

	//Use predefined special characters if user doesn't specify them
	specialCharSet := allowedSpecialChars
	if len(allowedSpecialChars) != 0 {
		specialCharSet = SpecialCharSet
	}

	//Set special character
	for i := 0; i < minSpecialChars; i++ {
		random := rand.Intn(len(specialCharSet))
		password.WriteString(string(specialCharSet[random]))
	}

	//Set numeric
	for i := 0; i < minNum; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	//Set uppercase
	for i := 0; i < minUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	remainingLength := passwordLength - minSpecialChars - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})

	return string(inRune)
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}

	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func parseNullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	} else {
		return ""
	}
}

func PairwiseReverseHex(number int64, parts int) string {
	result := ""

	for i := 0; i < parts; i++ {
		result += fmt.Sprintf("%02x", number%256)
		number = number / 256
	}

	return strings.ToUpper(result)
}

func AzureSIDToDatabaseSID(ctx context.Context, azureSid string) (databaseSID string) {
	databaseSID = "0x"
	parts := strings.Split(azureSid, "-")

	if len(parts) != 8 {
		logging.AddError(ctx, "SID format error", fmt.Sprintf("%s is not a valid Azure SID", azureSid))
		return
	}

	for i := 4; i <= 7; i++ {
		number, err := strconv.ParseInt(parts[i], 10, 64)
		if err != nil {
			logging.AddError(ctx, "SID format error", fmt.Sprintf("%s is not a valid Azure SID", azureSid))
			return
		}
		databaseSID += PairwiseReverseHex(number, 4)
	}

	return databaseSID
}

func ObjectIDToDatabaseSID(ctx context.Context, objectID string) (databaseSID string) {
	s := strings.ReplaceAll(objectID, "-", "")

	if len(s) != 32 {
		logging.AddError(ctx, "SID format error", fmt.Sprintf("%s is not a valid Object ID", objectID))
		return
	}

	databaseSID = "0x" + strings.ToUpper(s[6:8]+s[4:6]+s[2:4]+
		s[0:2]+s[10:12]+s[8:10]+s[14:16]+s[12:14]+
		s[16:18]+s[18:20]+s[20:22]+s[22:24]+s[24:26]+
		s[26:28]+s[28:30]+s[30:32])

	return databaseSID
}
