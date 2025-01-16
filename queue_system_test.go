package quest_system

import (
	"fmt"
	"testing"
	"time"
)

// TestQueueAndBankCounter 测试队列与银行柜台的功能
func TestQueueAndBankCounter(t *testing.T) {
	// 初始化排队队列
	queue := NewQueue()

	// 发放一些票号
	ticket1 := queue.IssueTicket("Alice", 1)
	t.Logf("New ticket issued: %d for customer %s", ticket1.Number, ticket1.Name)

	ticket2 := queue.IssueTicket("Bob", 3)
	t.Logf("New ticket issued: %d for customer %s", ticket2.Number, ticket2.Name)

	ticket3 := queue.IssueTicket("Charlie", 2)
	t.Logf("New ticket issued: %d for customer %s", ticket3.Number, ticket3.Name)

	// 发放相同优先级的票
	ticket4 := queue.IssueTicket("David", 3)
	t.Logf("New ticket issued: %d for customer %s", ticket4.Number, ticket4.Name)

	// 取消票号 ticket1
	if queue.CancelTicket(ticket1.Number) {
		t.Logf("Cancelled ticket %d", ticket1.Number)
	} else {
		t.Errorf("Failed to cancel ticket %d", ticket1.Number)
	}

	// 初始化银行柜台
	bankCounter := NewBankCounter(queue)

	// 模拟银行柜台并发服务
	bankCounter.wg.Add(3)

	mockFunc := func(ticket *Ticket) error { // 模拟服务过程
		// 计算排队时间
		waitTime := time.Since(ticket.QueueTime)

		fmt.Printf("Serving customer %s with ticket number %d. Wait time: %v\n", ticket.Name, ticket.Number, waitTime)
		time.Sleep(2 * time.Second) // 模拟服务时间
		fmt.Printf("Finished serving customer %s with ticket number %d\n", ticket.Name, ticket.Number)

		return nil
	}

	// 服务客户
	go func() {
		bankCounter.ServeCustomer(mockFunc)
	}()
	go func() {
		bankCounter.ServeCustomer(mockFunc)
	}()
	go func() {
		bankCounter.ServeCustomer(mockFunc)
	}()

	// 等待所有服务完成
	bankCounter.wg.Wait()

	// 尝试重置票号，队列为空，可以重置
	if !queue.ResetTicketNumber() {
		t.Errorf("Failed to reset ticket numbers")
	}
}
