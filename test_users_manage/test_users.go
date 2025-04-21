package testusersmanage

type TestUsersManager struct {
}

func (userManager *TestUsersManager) GetUserEmail(guid string) (string, error) {
	return "lyugaev2003@gmail.com", nil
}
