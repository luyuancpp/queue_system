package quest_system

import (
	"fmt"
	"sync"
)

// 定义服务函数类型
type ServeFunc func(ticket *Ticket) error

// BankCounter 代表银行柜台的服务
type BankCounter struct {
	queue *Queue
	wg    sync.WaitGroup
}

// NewBankCounter 创建一个新的银行柜台
func NewBankCounter(queue *Queue) *BankCounter {
	return &BankCounter{
		queue: queue,
	}
}

// ServeCustomer 服务客户，传入外部定义的服务函数
func (bc *BankCounter) ServeCustomer(serveFn ServeFunc) {
	// 获取一个客户的票号
	ticket, err := bc.queue.ServeTicket()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 启动一个新的 goroutine 来模拟服务过程
	defer bc.wg.Done() // 完成后减少计数器

	// 调用外部传入的服务函数
	if err := serveFn(ticket); err != nil {
		GetLogger().Info("Error serving customer %s with ticket number %d: %v\n", ticket.Name, ticket.Number, err)
	} else {
		GetLogger().Info("Finished serving customer %s with ticket number %d\n", ticket.Name, ticket.Number)
	}
}
