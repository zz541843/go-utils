package jz

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

/*
坑:
如果自动应答了,务必别手动应答

*/

type RabbitMQ struct {
	conn          *amqp.Connection
	ch            *amqp.Channel
	queue         *amqp.Queue
	QueueName     string //队列名称
	Exchange      string //交换机名称
	Key           string //bind Key 名称
	MqUrl         string //连接信息
	ChannelConfig *channelConfig
	//ChannelConfig map[string]string
	mode int
}

const (
	WorkQueue = iota + 1
	PublishQueue
	RoutingQueue
	TopicsQueue
)

type channelConfig struct {
	Qos struct {
		PrefetchSize  int
		PrefetchCount int
		Global        bool
	}
	Consume struct {
		QueueName   string
		ConsumerTag string
		NoLocal     bool
		AutoAck     bool
		Exclusive   bool
		NoWait      bool
		Arguments   amqp.Table
	}
	Publish struct {
		Exchange  string
		Key       string
		Mandatory bool
		Immediate bool
	}
}
type QueueDeclare struct {
	QueueName  string
	Durable    bool //是否持久化,队列在代理重新启动后仍然存在
	Exclusive  bool //是否独占,仅由一个连接使用，当该连接关闭时，队列将被删除
	AutoDelete bool //是否自动删除,最后一个消费者取消订阅时，删除至少有一个消费者的队列
	NoWait     bool //是否阻塞处理
	Arguments  amqp.Table
}

const MqUrl = "amqp://imoocuser:imoocuser@127.0.0.1:5672/imooc"

func (r *RabbitMQ) Destroy() {
	r.ch.Close()
	r.conn.Close()
}

//NewRabbitMQ 创建实例
func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	return &RabbitMQ{
		QueueName:     queueName,
		Exchange:      exchange,
		Key:           key,
		ChannelConfig: &channelConfig{},
	}
}
func (r *RabbitMQ) Dial(url string) (err error) {
	r.MqUrl = url
	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Println(4, err.Error())
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(5, err.Error())
		return err
	}
	q, err := ch.QueueDeclare(
		r.QueueName, // name
		true,        //是否持久化
		false,       //是否自动删除
		false,       //是否具有排他性
		false,       //是否阻塞处理
		nil,         //额外的属性
	)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	r.ch = ch
	r.queue = &q
	r.QueueName = q.Name

	return
}
func (r *RabbitMQ) Publish(msg string) (err error) {
	err = r.ch.Publish(
		r.ChannelConfig.Publish.Exchange,  // exchange
		r.ChannelConfig.Publish.Key,       // routing key
		r.ChannelConfig.Publish.Mandatory, // mandatory
		r.ChannelConfig.Publish.Immediate, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(fmt.Sprintf(msg)),
		})
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return
}
func (r *RabbitMQ) Consumer(f func(delivery *amqp.Delivery), a ...int64) (err error) {
	//qos := r.ChannelConfig.Qos
	consume := r.ChannelConfig.Consume
	//
	//err = r.ch.Qos(qos.PrefetchCount, qos.PrefetchSize, qos.Global)
	if err != nil {
		fmt.Println("qos error!")
		return err
	}

	msgs, err := r.ch.Consume(
		consume.QueueName,   // queue
		consume.ConsumerTag, // consumer
		consume.AutoAck,     // auto-ack
		consume.Exclusive,   // exclusive
		consume.NoLocal,     // no-local
		consume.NoWait,      // no-wait
		consume.Arguments,   // args
	)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	i := time.Second
	if len(a) == 0 {
		i = 0
	} else {
		i = time.Second * time.Duration(a[0])
	}
	go func() {
		for d := range msgs {
			f(&d)
			time.Sleep(i)
		}
	}()
	return
}

func (r *RabbitMQ) Work(url string, queueDeclare QueueDeclare) (*RabbitMQ, error) {
	r.mode = WorkQueue
	r.MqUrl = url
	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Println(4, err.Error())
		return r, err
	}
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(5, err.Error())
		return r, err
	}
	if len(queueDeclare.QueueName) == 0 {
		return nil, fmt.Errorf("QueueName can't nil")
	}
	q, err := ch.QueueDeclare(
		queueDeclare.QueueName,
		queueDeclare.Durable,
		queueDeclare.AutoDelete,
		queueDeclare.Exclusive,
		queueDeclare.NoWait,
		queueDeclare.Arguments,
	)
	if err != nil {
		fmt.Println(err.Error())
		return r, err
	}
	r.ch = ch
	r.queue = &q

	return r, nil
}

type ExchangeDeclare struct {
	ExchangeName string
	Type         string
	Durable      bool // 持久化
	AutoDelete   bool // 自动删除
	Internal     bool //
	NoWait       bool // 阻塞
	Arguments    amqp.Table
}

func (r *RabbitMQ) SubscribeProvider(url string, exchangeDeclare ExchangeDeclare) (*RabbitMQ, error) {

	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Println(4, err.Error())
		return r, err
	}
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(5, err.Error())
		return r, err
	}
	r.ch = ch

	err = r.ch.ExchangeDeclare(
		exchangeDeclare.ExchangeName,
		exchangeDeclare.Type,
		exchangeDeclare.Durable,
		exchangeDeclare.AutoDelete,
		exchangeDeclare.Internal,
		exchangeDeclare.NoWait,
		exchangeDeclare.Arguments,
	)

	if err != nil {
		fmt.Println(4, err.Error())
		return r, err
	}

	r.ChannelConfig.Publish.Exchange = exchangeDeclare.ExchangeName

	return r, nil
}

func (r *RabbitMQ) SubscribeConsumer(url string, exchangeDeclare ExchangeDeclare, queueDeclare QueueDeclare) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Println(4, err.Error())
		return r, err
	}
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(5, err.Error())
		return r, err
	}
	r.ch = ch
	err = r.ch.ExchangeDeclare(
		exchangeDeclare.ExchangeName,
		exchangeDeclare.Type,
		exchangeDeclare.Durable,
		exchangeDeclare.AutoDelete,
		exchangeDeclare.Internal,
		exchangeDeclare.NoWait,
		exchangeDeclare.Arguments,
	)

	if err != nil {
		fmt.Println(4, err.Error())
		return r, err
	}

	q, err := ch.QueueDeclare(
		queueDeclare.QueueName,
		queueDeclare.Durable,
		queueDeclare.AutoDelete,
		queueDeclare.Exclusive,
		queueDeclare.NoWait,
		queueDeclare.Arguments,
	)
	err = ch.QueueBind(
		q.Name,                       // queue name
		"",                           // routing key
		exchangeDeclare.ExchangeName, // exchange
		false,
		nil,
	)
	r.ChannelConfig.Consume.QueueName = q.Name
	r.ChannelConfig.Consume.AutoAck = true

	return r, nil
}
