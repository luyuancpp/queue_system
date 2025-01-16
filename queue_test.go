package quest_system

import (
	"math/rand"
	"testing"
	"time"
)

// TestQueueAndBankCounter 测试队列和银行柜台的功能
func TestQueueAndBankCounter(t *testing.T) {
	// 使用默认的日志记录器
	logger := &DefaultLogger{}
	SetLogger(logger)

	// 初始化排队队列
	queue := NewQueue()

	// 发放一些票号
	ticketAlice := queue.IssueTicket("Alice", 1)
	logger.Info("Issued ticket %d for customer %s", ticketAlice.Number, ticketAlice.Name)

	ticketBob := queue.IssueTicket("Bob", 3)
	logger.Info("Issued ticket %d for customer %s", ticketBob.Number, ticketBob.Name)

	ticketCharlie := queue.IssueTicket("Charlie", 2)
	logger.Info("Issued ticket %d for customer %s", ticketCharlie.Number, ticketCharlie.Name)

	// 发放相同优先级的票
	ticketDavid := queue.IssueTicket("David", 3)
	logger.Info("Issued ticket %d for customer %s", ticketDavid.Number, ticketDavid.Name)

	// 取消票号 ticketAlice
	if queue.CancelTicket(ticketAlice.Number) {
		logger.Info("Cancelled ticket %d", ticketAlice.Number)
	} else {
		logger.Error("Failed to cancel ticket %d", ticketAlice.Number)
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
		logger.Info("Serving customer %s with ticket number %d. Wait time: %v", ticket.Name, ticket.Number, waitTime)

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
		logger.Error("Failed to reset ticket numbers")
	}
}

// 测试 ticketIndexMap 和 tickets 的一致性，并随机取消一些票
func TestTicketIndexMapConsistencyWithCancellation(t *testing.T) {
	// 自定义票的数量，默认是10万张
	numTickets := 100000
	q := NewQueue()

	// 发放 numTickets 张票
	for i := 0; i < numTickets; i++ {
		q.IssueTicket("Customer", uint32(i))
	}

	// 随机取消一些票，假设取消总数为总票数的 10%
	cancelCount := numTickets / 10
	rand.Seed(time.Now().UnixNano()) // 初始化随机数种子

	// 随机取消票
	for i := 0; i < cancelCount; i++ {
		// 随机选择一个票号并取消
		randomTicketIndex := rand.Intn(numTickets)
		ticket := q.tickets[randomTicketIndex]
		q.CancelTicket(ticket.Number)
	}

	// 验证 ticketIndexMap 和 tickets 的一致性
	checkTicketConsistency(t, q)
}

// 检查 ticketIndexMap 和 tickets 的一致性
func checkTicketConsistency(t *testing.T, q *Queue) {
	// 确保 ticketIndexMap 中每个票号都能正确指向 tickets 数组中的位置
	for i, ticket := range q.tickets {
		if ticket.IsCancelled {
			// 如果票被取消，跳过此票的验证
			continue
		}
		if index, exists := q.ticketIndexMap[ticket.Number]; exists {
			if index != i {
				t.Errorf("Ticket %d in ticketIndexMap points to wrong index: expected %d, got %d", ticket.Number, i, index)
			}
		} else {
			t.Errorf("Ticket %d is missing from ticketIndexMap", ticket.Number)
		}
	}

	// 确保 ticketIndexMap 中的每个索引都有对应的票号
	for ticketNumber, index := range q.ticketIndexMap {
		if index >= len(q.tickets) || q.tickets[index].Number != ticketNumber {
			t.Errorf("ticketIndexMap has incorrect mapping: ticketNumber %d maps to index %d, but tickets[%d].Number = %d", ticketNumber, index, index, q.tickets[index].Number)
		}
	}
}
