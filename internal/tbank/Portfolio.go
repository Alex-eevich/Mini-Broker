package tbank

type Account struct {
	Accounts []struct {
		Id string `json:"id"`
	} `json:"accounts"`
}

type SandboxPortfolio struct {
	AccountId string `json:"accountId"`
	Currency  string `json:"currency"`
}

func (c *Client) GetListAccounts() ([]string, error) {
	var account Account

	err := c.do(
		"POST",
		"tinkoff.public.invest.api.contract.v1.UsersService/GetAccounts",
		map[string]string{},
		&account,
	)
	if err != nil {
		return nil, err
	}

	var accounts []string
	for _, account := range account.Accounts {
		accounts = append(accounts, account.Id)
	}

	return accounts, nil
}
