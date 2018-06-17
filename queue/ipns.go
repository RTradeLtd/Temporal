package queue

/*
func ProcessIPNSPublishRequests(msgs <-chan amqp.Delivery, db *gorm.DB) error {
	var ipnsUpdate IPNSUpdate
	var resolve bool
	var switchErr bool
	rtfs, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}

	for d := range msgs {
		err := json.Unmarshal(d.Body, &ipnsUpdate)
		if err != nil {
			// TODO: handle
			fmt.Println("error unmarshaling into ipns update struct ", err)
			d.Ack(false)
			continue
		}
		contentHash := ipnsUpdate.CID
		ttl := ipnsUpdate.TTL
		key := ipnsUpdate.Key
		lifetime := ipnsUpdate.LifeTime
		resolveStr := ipnsUpdate.Resolve
		switch resolveStr {
		case "true":
			resolve = true
		case "false":
			resolve = false
		default:
			// TODO: handle
			fmt.Println("errror, resolve is neither \"true\" or \"false\" ")
			switchErr = true
		}
		if switchErr {
			// TODO: handle
			fmt.Println("errror, resolve is neither \"true\" or \"false\" ")
			d.Ack(false)
			continue
		}

		resp, err := rtfs.PublishToIPNSDetails(contentHash, lifetime, ttl, key, resolve)
		if err != nil {
			// TODO: handle
			fmt.Println("error publishing to ipns ", err)
			d.Ack(false)
			continue
		}
		fmt.Println("record published")
		fmt.Printf("name: %s\nvalue: %s\n", resp.Name, resp.Value)
		d.Ack(false)
	}
	return nil
}
*/
