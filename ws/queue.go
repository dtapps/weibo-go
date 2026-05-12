package ws

import "fmt"

// sendTask 有序发送任务（闭包形式）
type sendTask struct {
	execute func() error    // 实际的发送逻辑
	result  chan sendResult // 结果回传通道
}

// sendResult 发送结果
type sendResult struct {
	err error
}

// ErrSendQueueFull 发送队列已满
var ErrSendQueueFull = fmt.Errorf("发送队列已满")

// startSender 启动有序发送协程（懒加载，只启动一次）
// 所有 SendMessage 的请求都会进入队列，
// 由该协程按 FIFO 顺序逐条执行，保证消息有序性。
func (c *WsClient) startSender() {
	c.sendOnce.Do(func() {
		go c.senderLoop()
	})
}

// senderLoop 有序发送主循环
// 从 sendQueue 中按序取出任务并执行，保证消息严格按入队顺序发出。
func (c *WsClient) senderLoop() {
	c.log.Info("[Sender] 有序发送协程已启动")
	for {
		select {
		case <-c.senderDone:
			// 客户端关闭，退出循环，排空剩余任务
			for len(c.sendQueue) > 0 {
				task := <-c.sendQueue
				task.result <- sendResult{err: fmt.Errorf("客户端已关闭")}
			}
			c.log.Info("[Sender] 有序发送协程已停止")
			return

		case task := <-c.sendQueue:
			err := task.execute()
			task.result <- sendResult{err: err}
		}
	}
}

// SendQueued 通过有序队列发送（闭包模式）
func (c *WsClient) SendQueued(execute func() error) error {
	c.startSender()

	resultCh := make(chan sendResult, 1)
	task := sendTask{
		execute: execute,
		result:  resultCh,
	}

	select {
	case c.sendQueue <- task:
	default:
		return ErrSendQueueFull
	}

	result := <-resultCh
	return result.err
}
