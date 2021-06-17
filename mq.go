package Mq

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup
type MQ interface{
	publish()
	subscribe()
	unsubscribe()
	close()
}
type broker struct{
	exit chan bool
	capacity int
	topics map[string][]chan interface{}
	sync.RWMutex
}
func(a *broker )subscribe(topic string)(<-chan interface{}, error){
	select{
		case <-a.exit:
			return nil,errors.New("closed!")
			default:
	}
	ch := make(chan interface{},a.capacity)
	a.Lock()
	a.topics[topic]=append(a.topics[topic],ch)
	a.Unlock()
	return ch,nil
}
func(a *broker)unsubscribe(topic string,target <-chan interface{})error{
	select{
	case <-a.exit:
		return errors.New("closed!")
	default:
	}
	a.RLock()
	subscribers:=a.topics[topic]
	a.RUnlock()
	var newSubs []chan interface{}
	for _,subscriber :=range subscribers{
		if subscriber==target{
			continue
		}
		newSubs=append(newSubs,subscriber)
	}
	a.Lock()
	a.topics[topic]=newSubs
	a.Unlock()
	return nil
}
func(a *broker)publish(topic string, msg interface{}) error{
	select {
	case <-a.exit:
		return errors.New("broker closed")
	default:
	}
	a.RLock()
	subscribers, ok := a.topics[topic]
	a.RUnlock()
	if !ok {
		return nil
	}
	num:= len(subscribers)
	for i:=0;i<num;i++{
		select {
		case subscribers[i]<-msg:
		case <-time.After(time.Millisecond * 5):
		case <-a.exit:
		}
	}
	return nil
}

func (a *broker)close(){
	select {
	case <-a.exit:
		return
	default:
		close(a.exit)
		a.Lock()
		a.topics = make(map[string][]chan interface{})
		a.Unlock()
	}
	return
}
func (a *broker)getMsg(sub <-chan interface{}){
	for msg:=range sub{
		if msg!=nil{
			fmt.Printf("接受的消息为：%s\n",msg)
		}
	}
}

func pushMsg(topic string,a *broker){
	defer wg.Done()
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	i:=1
	for i<10{
		select {
		case <- t.C:
			msg:=fmt.Sprintf("%s:第%d天新闻联播",topic,i)
			i++
			err := a.publish(topic,msg)
			fmt.Printf("成功推送%s消息\n",topic)
			if err != nil{
				fmt.Println("pub message failed")
			}
		default:
		}
	}
}
func test(topic string,a *broker){
	ch1,err:=a.subscribe(topic)
	if err!=nil{
		fmt.Println("subscribe failed")
	}
	go pushMsg(topic,a)
	a.getMsg(ch1)
	defer a.close()
}
//func main(){
//	a:=&broker{
//		topics:   make(map[string][]chan interface{}),
//		capacity: 10,
//	}
//	wg.Add(2)
//	go test("topic1",a)
//	go test("topic2",a)
//	wg.Wait()
//
//}
