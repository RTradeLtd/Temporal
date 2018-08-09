package queue

import (
	"github.com/RTradeLtd/Temporal/config"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

func (qm *QueueManager) ProcessIPFSClusterAdds(msgs <-chan amqp.Delivery, cfg *config.TemporalConfig, db *gorm.DB) error {
	return nil
}
