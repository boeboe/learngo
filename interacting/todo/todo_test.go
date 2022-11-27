package todo_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/boeboe/learngo/interacting/todo"
)

func TestAdd(t *testing.T) {
	l := todo.List{}

	taskName := "New task"
	l.Add(taskName)

	if l[0].Task != taskName {
		t.Errorf("expected %q, got %q instead", taskName, l[0].Task)
	}
}
func TestComplete(t *testing.T) {
	l := todo.List{}

	taskName := "New task"
	l.Add(taskName)

	if l[0].Task != taskName {
		t.Errorf("expected %q, got %q instead", taskName, l[0].Task)
	}

	if l[0].Done {
		t.Errorf("new task should not be completed")
	}
	l.Complete(1)
	if !l[0].Done {
		t.Errorf("new task should be completed")
	}
}

func TestDelete(t *testing.T) {
	l := todo.List{}
	tasks := []string{
		"New task 1",
		"New task 2",
		"New taks 3",
	}
	for _, v := range tasks {
		l.Add(v)
	}

	if l[0].Task != tasks[0] {
		t.Errorf("expected %q, got %q instead", tasks[0], l[0].Task)
	}

	l.Delete(2)
	if len(l) != 2 {
		t.Errorf("expected list length %d, got %d instead", 2, len(l))
	}
	if l[1].Task != tasks[2] {
		t.Errorf("expected %q, got %q instead", tasks[2], l[1].Task)
	}
}

func TestSaveGet(t *testing.T) {
	l1 := todo.List{}
	l2 := todo.List{}

	taskName := "New task"
	l1.Add(taskName)

	if l1[0].Task != taskName {
		t.Errorf("expected %q, got %q instead", taskName, l1[0].Task)
	}

	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Errorf("error creating temp file: %s", err)
	}

	defer os.Remove(tf.Name())

	if err := l1.Save(tf.Name()); err != nil {
		t.Fatalf("error saving list to file: %s", err)
	}
	if err := l2.Get(tf.Name()); err != nil {
		t.Fatalf("error getting list from file: %s", err)
	}
	if l1[0].Task != l2[0].Task {
		t.Errorf("task %q should match task %q", l1[0].Task, l2[0].Task)
	}

}
