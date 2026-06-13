package ssh

func (c *Client) Exec(cmd string) (string, error) {

	out, errstr, _, err := c.Run(cmd)
	if err != nil {
		return errstr, err
	}
	if errstr != "" {
		return errstr, nil
	}
	return out, nil
}
