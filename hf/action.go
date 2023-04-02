package hf

import "errors"

type manipulatorImpl struct {
	myClient  Client
	structure structure
}

type Manipulator interface {
	SetClient(client Client) (bool, error)
	Find(label string) *manipulatorImpl
	Get() (*Component, error)
	Set(value interface{}) error
	FindAndSet(pos []string, value interface{}) error
	FindAndGet(pos []string) (*Component, error)
	FindAndExecute(pos []string, linkMode int) (bool, error)
	Execute(linkMode int) (bool, error)
}

func NewManipulator() Manipulator {
	return &manipulatorImpl{}
}

func NewManipulatorWithClient(client Client) Manipulator {
	return &manipulatorImpl{
		myClient: client,
		structure: structure{
			elem:     nil,
			children: client.GetAppStructure(),
		},
	}
}

func (m *manipulatorImpl) SetClient(client Client) (bool, error) {
	if client == nil {
		return false, errors.New("client 不可为空")
	}
	m.myClient = client
	m.structure.children = client.GetAppStructure()
	return true, nil
}

// Find 选择最长匹配的合法坐标
func (m *manipulatorImpl) Find(label string) *manipulatorImpl {
	if m.structure.children == nil {
		return m
	}
	s, ok := m.structure.children[label]
	if !ok {
		return m
	}
	return &manipulatorImpl{
		myClient:  m.myClient,
		structure: s,
	}
}

func (m *manipulatorImpl) Get() (*Component, error) {
	if m.structure.elem == nil {
		return nil, errors.New("没有匹配的模块")
	}
	return m.structure.elem, nil
}

func (m *manipulatorImpl) Set(value interface{}) error {
	if m.structure.elem == nil {
		return errors.New("没有匹配的模块")
	}
	m.myClient.SetData(m.structure.elem.ID, value)
	return nil
}

func (m *manipulatorImpl) Execute(linkMode int) (bool, error) {
	if m.structure.elem == nil {
		return false, errors.New("没有匹配的模块")
	}
	c := m.myClient
	err := c.InvokeFunc(m.structure.elem.ID, linkMode)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *manipulatorImpl) FindAndSet(pos []string, value interface{}) (err error) {
	mCopy := m
	for _, v := range pos {
		mCopy = mCopy.Find(v)
	}
	err = mCopy.Set(value)
	return
}
func (m *manipulatorImpl) FindAndGet(pos []string) (component *Component, err error) {
	mCopy := m
	for _, v := range pos {
		mCopy = mCopy.Find(v)
	}
	component, err = mCopy.Get()
	return
}
func (m *manipulatorImpl) FindAndExecute(pos []string, linkMode int) (ok bool, err error) {
	mCopy := m
	for _, v := range pos {
		mCopy = mCopy.Find(v)
	}
	ok, err = mCopy.Execute(linkMode)
	return
}
