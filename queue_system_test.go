package quest_system

import (
	"testing"
	"time"
)

// TestQueueAndBankCounter 测试队列和银行柜台的功能
func TestQueueAndBankCounter(t *testing.T) {
	// 初始化排队队列
	queue := NewQueue()

	// 发放一些票号
	ticketAlice := queue.IssueTicket("Alice", 1)
	t.Logf("Issued ticket %d for customer %s", ticketAlice.Number, ticketAlice.Name)

	ticketBob := queue.IssueTicket("Bob", 3)
	t.Logf("Issued ticket %d for customer %s", ticketBob.Number, ticketBob.Name)

	ticketCharlie := queue.IssueTicket("Charlie", 2)
	t.Logf("Issued ticket %d for customer %s", ticketCharlie.Number, ticketCharlie.Name)

	// 发放相同优先级的票
	ticketDavid := queue.IssueTicket("David", 3)
	t.Logf("Issued ticket %d for customer %s", ticketDavid.Number, ticketDavid.Name)

	// 取消票号 ticketAlice
	if queue.CancelTicket(ticketAlice.Number) {
		t.Logf("Cancelled ticket %d", ticketAlice.Number)
	} else {
		t.Errorf("Failed to cancel ticket %d", ticketAlice.Number)
	}

	// 初始化银行柜台
	bankCounter := NewBankCounter(queue)

	// 模拟银行柜台并发服务
	bankCounter.wg.Add(3)

	// 模拟服务过程
	serveCustomer := func(ticket *Ticket) error {
		// 计算排队时间
		waitTime := time.Since(ticket.QueueTime)

		// 打印服务信息
		t.Logf("Serving customer %s with ticket number %d. Wait time: %v", ticket.Name, ticket.Number, waitTime)

		return nil
	}

	// 并发服务客户
	go func() {
		bankCounter.ServeCustomer(serveCustomer)
	}()
	go func() {
		bankCounter.ServeCustomer(serveCustomer)
	}()
	go func() {
		bankCounter.ServeCustomer(serveCustomer)
	}()

	// 等待所有服务完成
	bankCounter.wg.Wait()

	// 尝试重置票号
	if !queue.ResetTicketNumber() {
		t.Errorf("Failed to reset ticket numbers")
	}
}
