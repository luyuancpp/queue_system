package main

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

const ticketExpirationDuration = 30 * time.Second // 票号有效期为30秒

// Ticket 代表一个客户的票号
type Ticket struct {
	Number    uint32
	Name      string
	QueueTime time.Time // 客户排队的时间
	IsExpired bool
	Priority  uint32 // 用于优先队列的优先级
}

// Queue 代表排队的队列，使用优先队列（堆）实现
type Queue struct {
	tickets        []*Ticket
	nextTicketNum  uint32 // 记录下一个生成的票号
	mu             sync.Mutex
	expirationTime time.Duration
	ticketIndexMap map[uint32]int // 用于快速查找票号在队列中的位置
}

// NewQueue 创建一个空的排队队列
func NewQueue(expirationTime time.Duration) *Queue {
	return &Queue{
		tickets:        make([]*Ticket, 0),
		nextTicketNum:  0,
		expirationTime: expirationTime,
		ticketIndexMap: make(map[uint32]int),
	}
}

// IssueTicket 发放一个新的票号
func (q *Queue) IssueTicket(name string, priority uint32) *Ticket {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 生成一个新的票号
	ticket := &Ticket{
		Number:    q.nextTicketNum,
		Name:      name,
		QueueTime: time.Now(),
		IsExpired: false,
		Priority:  priority,
	}

	// 将票号加入优先队列
	heap.Push(q, ticket)

	// 更新下一个生成的票号
	q.nextTicketNum++

	// 将票号的索引保存到映射中
	q.ticketIndexMap[ticket.Number] = len(q.tickets) - 1

	return ticket
}

// CancelTicket 取消指定票号的客户
func (q *Queue) CancelTicket(ticketNumber uint32) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 查找并取消该票号
	if index, exists := q.ticketIndexMap[ticketNumber]; exists {
		// 用最后一个元素替换当前元素并移除最后一个元素
		lastTicket := q.tickets[len(q.tickets)-1]
		q.tickets[index] = lastTicket
		q.ticketIndexMap[lastTicket.Number] = index

		// 删除最后一个元素
		q.tickets = q.tickets[:len(q.tickets)-1]
		delete(q.ticketIndexMap, ticketNumber)

		// 调整堆
		heap.Fix(q, index)

		return true
	}

	return false // 未找到票号
}

// ServeTicket 服务队列中的下一个客户
func (q *Queue) ServeTicket() (*Ticket, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 取出队列中的第一个客户进行服务
	if len(q.tickets) == 0 {
		return nil, fmt.Errorf("no customers in queue")
	}

	// 获取优先队列中优先级最高的票
	ticket := heap.Pop(q).(*Ticket)
	delete(q.ticketIndexMap, ticket.Number) // 从映射中删除该票号的索引

	return ticket, nil
}

// ExpireTickets 检查并过期无效的票号
func (q *Queue) ExpireTickets() {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 过期队列中的票号
	for _, t := range q.tickets {
		if time.Since(t.QueueTime) > q.expirationTime {
			t.IsExpired = true
		}
	}
}

// ResetTicketNumber 重置票号计数器，从0开始，仅当队列为空时才重置
func (q *Queue) ResetTicketNumber() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tickets) > 0 {
		// 如果队列中还有客户，则不能重置票号
		fmt.Println("Cannot reset ticket numbers, queue is not empty.")
		return false
	}

	// 清空队列并重置票号计数器
	q.nextTicketNum = 0
	q.tickets = nil
	fmt.Println("Ticket numbers have been reset.")
	return true
}

// GetQueueSize 返回当前排队中的人数
func (q *Queue) GetQueueSize() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.tickets)
}

// 获取指定票号的索引位置
func (q *Queue) GetTicketIndex(ticketNumber uint32) (int, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	index, exists := q.ticketIndexMap[ticketNumber]
	return index, exists
}

// 定义堆实现（优先队列）
func (q *Queue) Len() int {
	return len(q.tickets)
}

func (q *Queue) Less(i, j int) bool {
	return q.tickets[i].Priority > q.tickets[j].Priority // 优先级较高的在前
}

func (q *Queue) Swap(i, j int) {
	q.tickets[i], q.tickets[j] = q.tickets[j], q.tickets[i]
	q.ticketIndexMap[q.tickets[i].Number] = i
	q.ticketIndexMap[q.tickets[j].Number] = j
}

func (q *Queue) Push(x interface{}) {
	ticket := x.(*Ticket)
	q.tickets = append(q.tickets, ticket)
	q.ticketIndexMap[ticket.Number] = len(q.tickets) - 1
}

func (q *Queue) Pop() interface{} {
	old := q.tickets
	n := len(old)
	ticket := old[n-1]
	q.tickets = old[0 : n-1]
	delete(q.ticketIndexMap, ticket.Number)
	return ticket
}

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

// ServeCustomer 服务客户
func (bc *BankCounter) ServeCustomer() {
	defer bc.wg.Done()

	// 获取一个客户的票号
	ticket, err := bc.queue.ServeTicket()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 如果票号过期，跳过服务
	if ticket.IsExpired {
		fmt.Printf("Ticket %d has expired, skipping service\n", ticket.Number)
		return
	}

	// 计算排队时间
	waitTime := time.Since(ticket.QueueTime)

	// 模拟服务过程
	fmt.Printf("Serving customer %s with ticket number %d. Wait time: %v\n", ticket.Name, ticket.Number, waitTime)
	time.Sleep(2 * time.Second) // 模拟服务时间
	fmt.Printf("Finished serving customer %s with ticket number %d\n", ticket.Name, ticket.Number)
}

func main() {
	// 初始化排队队列，票号有效期为30秒
	queue := NewQueue(ticketExpirationDuration)

	// 发放一些票号
	ticket1 := queue.IssueTicket("Alice", 1)
	fmt.Printf("New ticket issued: %d for customer %s\n", ticket1.Number, ticket1.Name)

	ticket2 := queue.IssueTicket("Bob", 3)
	fmt.Printf("New ticket issued: %d for customer %s\n", ticket2.Number, ticket2.Name)

	ticket3 := queue.IssueTicket("Charlie", 2)
	fmt.Printf("New ticket issued: %d for customer %s\n", ticket3.Number, ticket3.Name)

	// 显示当前排队人数
	fmt.Printf("Current queue size: %d\n", queue.GetQueueSize())

	// 获取指定票号的位置
	index, found := queue.GetTicketIndex(ticket2.Number)
	if found {
		fmt.Printf("Ticket %d is at position %d in the queue\n", ticket2.Number, index)
	}

	// 取消票号 ticket1
	if queue.CancelTicket(ticket1.Number) {
		fmt.Printf("Cancelled ticket %d\n", ticket1.Number)
	}

	// 显示取消后的排队人数
	fmt.Printf("Queue size after canceling ticket %d: %d\n", ticket1.Number, queue.GetQueueSize())

	// 初始化银行柜台
	bankCounter := NewBankCounter(queue)

	// 模拟银行柜台并发服务
	bankCounter.wg.Add(3)

	// 服务客户
	go bankCounter.ServeCustomer() // 服务 Bob (ticket2)
	go bankCounter.ServeCustomer() // 服务 Charlie (ticket3)

	// 定时检查过期票号
	go func() {
		for {
			time.Sleep(5 * time.Second)
			queue.ExpireTickets()
		}
	}()

	// 等待所有服务完成
	bankCounter.wg.Wait()

	// 尝试重置票号，队列为空，可以重置
	queue.ResetTicketNumber() // 应该成功重置
}
