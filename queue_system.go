package main

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

// Ticket 代表一个客户的票号
type Ticket struct {
	Number      uint32
	Name        string
	QueueTime   time.Time // 客户排队的时间
	Priority    uint32    // 用于优先队列的优先级
	CreatedAt   time.Time // 记录票的创建时间，用于处理优先级相同的情况
	IsCancelled bool      // 标记票是否被取消
}

// Queue 代表排队的队列，使用优先队列（堆）实现
type Queue struct {
	tickets        []*Ticket
	nextTicketNum  uint32 // 记录下一个生成的票号
	mu             sync.Mutex
	ticketIndexMap map[uint32]int // 用于快速查找票号在队列中的位置
}

func NewQueue() *Queue {
	return &Queue{
		tickets:        make([]*Ticket, 0),
		nextTicketNum:  0,
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
		Priority:  priority,
		CreatedAt: time.Now(), // 记录创建时间
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
		q.tickets[index].IsCancelled = true // 标记为取消
		// 不需要调整堆，取消标记后，堆会在取票时自动跳过已取消的票
		return true
	}

	return false // 未找到票号
}

// ServeTicket 服务队列中的下一个客户
func (q *Queue) ServeTicket() (*Ticket, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 取出队列中的第一个客户进行服务
	for len(q.tickets) > 0 {
		ticket := heap.Pop(q).(*Ticket)
		delete(q.ticketIndexMap, ticket.Number)

		// 如果票被取消，跳过此票
		if ticket.IsCancelled {
			fmt.Printf("Ticket %d is cancelled, skipping\n", ticket.Number)
			continue
		}

		// 返回有效的票
		return ticket, nil
	}

	return nil, fmt.Errorf("no customers in queue")
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

// 修改比较规则，首先按优先级排序，优先级相同则按创建时间排序
func (q *Queue) Less(i, j int) bool {
	if q.tickets[i].Priority == q.tickets[j].Priority {
		// 如果优先级相同，按创建时间排序
		return q.tickets[i].CreatedAt.Before(q.tickets[j].CreatedAt)
	}
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

	// 计算排队时间
	waitTime := time.Since(ticket.QueueTime)

	// 模拟服务过程
	fmt.Printf("Serving customer %s with ticket number %d. Wait time: %v\n", ticket.Name, ticket.Number, waitTime)
	time.Sleep(2 * time.Second) // 模拟服务时间
	fmt.Printf("Finished serving customer %s with ticket number %d\n", ticket.Name, ticket.Number)
}

func main() {
	// 初始化排队队列
	queue := NewQueue()

	// 发放一些票号
	ticket1 := queue.IssueTicket("Alice", 1)
	fmt.Printf("New ticket issued: %d for customer %s\n", ticket1.Number, ticket1.Name)

	ticket2 := queue.IssueTicket("Bob", 3)
	fmt.Printf("New ticket issued: %d for customer %s\n", ticket2.Number, ticket2.Name)

	ticket3 := queue.IssueTicket("Charlie", 2)
	fmt.Printf("New ticket issued: %d for customer %s\n", ticket3.Number, ticket3.Name)

	// 发放相同优先级的票
	ticket4 := queue.IssueTicket("David", 3)
	fmt.Printf("New ticket issued: %d for customer %s\n", ticket4.Number, ticket4.Name)

	// 取消票号 ticket1
	if queue.CancelTicket(ticket1.Number) {
		fmt.Printf("Cancelled ticket %d\n", ticket1.Number)
	}

	// 初始化银行柜台
	bankCounter := NewBankCounter(queue)

	// 模拟银行柜台并发服务
	bankCounter.wg.Add(3)

	// 服务客户
	go bankCounter.ServeCustomer() // 服务 Bob (ticket2)
	go bankCounter.ServeCustomer() // 服务 Charlie (ticket3)
	go bankCounter.ServeCustomer() // 服务 David (ticket4)

	// 等待所有服务完成
	bankCounter.wg.Wait()

	// 尝试重置票号，队列为空，可以重置
	queue.ResetTicketNumber() // 应该成功重置
}
