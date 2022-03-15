package tests

import (
	"fmt"
	"github.com/streadway/amqp"
	. "github.com/zz541843/go-utils"
	"time"
)

var mqUrl = "amqp://guest:guest@192.168.124.129:5672/"

func main() {
	test3()
	forever := make(chan bool)
	<-forever
}
func test3() {
	go func() {
		subscribeProvider, err := NewRabbitMQ("", "", "").
			SubscribeProvider(mqUrl, ExchangeDeclare{
				ExchangeName: "e1",
				Type:         amqp.ExchangeFanout,
				Durable:      true,
			})
		if err != nil {
			return
		}
		for i := 0; i < 111; i++ {
			err := subscribeProvider.Publish(fmt.Sprintf("%d", i))
			if err != nil {
				fmt.Println(2, err.Error())
				return
			}
			fmt.Println("已发送", i)
			time.Sleep(time.Second)
		}
	}()
	go func() {
		work, err := NewRabbitMQ("", "", "").
			SubscribeConsumer(
				mqUrl,
				ExchangeDeclare{
					ExchangeName: "e1",
					Type:         amqp.ExchangeFanout,
					Durable:      true,
				},
				QueueDeclare{
					Exclusive: true,
				})
		if err != nil {
			fmt.Println(5, err.Error())
			return
		}
		//work.ChannelConfig.Qos.PrefetchCount = 1
		err = work.Consumer(func(delivery *amqp.Delivery) {
			fmt.Println("1消费了：", string(delivery.Body))
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()
	go func() {
		work, err := NewRabbitMQ("", "", "").
			SubscribeConsumer(
				mqUrl,
				ExchangeDeclare{
					ExchangeName: "e1",
					Type:         amqp.ExchangeFanout,
					Durable:      true,
				},
				QueueDeclare{
					Exclusive: true,
				})
		if err != nil {
			fmt.Println(5, err.Error())
			return
		}
		//work.ChannelConfig.Qos.PrefetchCount = 1
		err = work.Consumer(func(delivery *amqp.Delivery) {
			fmt.Println("2消费了：", string(delivery.Body))
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()
}
func test2() {
	go func() {
		work, err := NewRabbitMQ("", "", "").Work(mqUrl, QueueDeclare{
			QueueName: "q1",
		})
		if err != nil {
			return
		}
		for i := 0; i < 111; i++ {
			err := work.Publish(fmt.Sprintf("%d", i))
			if err != nil {
				fmt.Println(2, err.Error())
				return
			}
			fmt.Println("已发送", i)
			time.Sleep(time.Second)
		}
	}()
	go func() {
		fmt.Println(1)
		work, err := NewRabbitMQ("", "", "").Work(mqUrl, QueueDeclare{
			QueueName: "q1",
		})
		if err != nil {
			fmt.Println(5, err.Error())
			return
		}
		work.ChannelConfig.Qos.PrefetchCount = 1
		err = work.Consumer(func(delivery *amqp.Delivery) {
			fmt.Println("1消费了：", string(delivery.Body))
			delivery.Ack(true)
		}, 22)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()
	go func() {
		fmt.Println(1)
		work, err := NewRabbitMQ("", "", "").Work(mqUrl, QueueDeclare{
			QueueName: "q1",
		})
		if err != nil {
			fmt.Println(5, err.Error())
			return
		}
		work.ChannelConfig.Qos.PrefetchCount = 1
		err = work.Consumer(func(delivery *amqp.Delivery) {
			fmt.Println("2消费了：", string(delivery.Body))
			delivery.Ack(true)
		}, 1)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()
}

func test1() {
	go func() {
		r := NewRabbitMQ("q1", "", "")
		err := r.Dial(mqUrl)
		if err != nil {
			fmt.Println(1, err.Error())
			return
		}
		for i := 0; i < 111; i++ {
			err := r.Publish(fmt.Sprintf("%d", i))
			if err != nil {
				fmt.Println(2, err.Error())
				return
			}
			fmt.Println("已发送", i)
			time.Sleep(time.Second)
		}
	}()
	go func() {
		r := NewRabbitMQ("q1", "", "")
		r.ChannelConfig.Qos.PrefetchCount = 1
		err := r.Dial(mqUrl)
		if err != nil {
			fmt.Println(1, err.Error())
			return
		}
		err = r.Consumer(func(delivery *amqp.Delivery) {
			fmt.Println("1消费了：", string(delivery.Body))
			delivery.Ack(true)
		}, 30)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()
	go func() {
		r := NewRabbitMQ("q1", "", "")
		r.ChannelConfig.Qos.PrefetchCount = 1
		err := r.Dial(mqUrl)
		if err != nil {
			fmt.Println(1, err.Error())
			return
		}
		err = r.Consumer(func(delivery *amqp.Delivery) {
			fmt.Println("2消费了：", string(delivery.Body))
			delivery.Ack(true)
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()
	forever := make(chan bool)
	<-forever

}
