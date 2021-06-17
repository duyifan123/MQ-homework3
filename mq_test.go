package Mq
import "testing"
func TestMq(t *testing.T){
	a:=&broker{
		topics:   make(map[string][]chan interface{}),
		capacity: 10,
	}
	wg.Add(2)
	go test("topic1",a)
	go test("topic2",a)
	wg.Wait()
}
