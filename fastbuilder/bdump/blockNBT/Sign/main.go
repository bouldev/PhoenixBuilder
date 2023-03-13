package blockNBT_Sign

import "fmt"

// 放置告示牌并写入告示牌数据的主函数
func (s *Sign) Main() error {
	err := s.Decode()
	if err != nil {
		return fmt.Errorf("Main: %v", err)
	}
	err = s.WriteDatas()
	if err != nil {
		return fmt.Errorf("Main: %v", err)
	}
	return nil
}
