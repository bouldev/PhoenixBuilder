package NBTAssigner

import (
	"fmt"
	GameInterface "phoenixbuilder/game_control/game_interface"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 检查 b.ItemPackage.Item.Custom.ItemTag 是否可以仅使用命令生成
func (b *Book) SpecialCheck() (bool, error) {
	err := b.Decode()
	if err != nil {
		return false, fmt.Errorf("SpecialCheck: %v", err)
	}
	b.ItemPackage.AdditionalData.Decoded = true
	// 解码
	if len(b.BookData.Author) == 0 && len(b.BookData.Pages) == 0 && len(b.BookData.Title) == 0 {
		return false, nil
	}
	return true, nil
	// 判断并返回值
}

// 从 b.ItemPackage.Item.Custom.ItemTag 提取成书数据，
// 然后保存在 b.BookData 中
func (b *Book) Decode() error {
	var pages []string = []string{}
	var author string = ""
	var title string = ""
	tag := b.ItemPackage.Item.Custom.ItemTag
	// 初始化
	if pages_origin, ok := tag["pages"]; ok {
		pages_got, success := pages_origin.([]interface{})
		if !success {
			return fmt.Errorf("Decode: Failed to convert pages_origin into []interface{}; tag = %#v", tag)
		}
		for key, value := range pages_got {
			page, success := value.(map[string]interface{})
			if !success {
				return fmt.Errorf("Decode: Failed to convert pages_got[%d] into map[string]interface{}; tag = %#v", key, tag)
			}
			text_origin, ok := page["text"]
			if !ok {
				continue
			}
			text_got, success := text_origin.(string)
			if !success {
				return fmt.Errorf(`Decode: Failed to convert pages_got[%d]["text"] into string; tag = %#v`, key, tag)
			}
			pages = append(pages, text_got)
		}
	}
	// pages
	if author_origin, ok := tag["author"]; ok {
		author_got, success := author_origin.(string)
		if !success {
			return fmt.Errorf("Decode: Failed to convert author_origin into string; tag = %#v", tag)
		}
		author = author_got
	}
	// author
	if title_origin, ok := tag["title"]; ok {
		title_got, success := title_origin.(string)
		if !success {
			return fmt.Errorf("Decode: Failed to convert title_origin into string; tag = %#v", tag)
		}
		title = title_got
	}
	// title
	b.BookData = BookData{
		Pages:  pages,
		Author: author,
		Title:  title,
	}
	return nil
	// return
}

func (b *Book) WriteData() error {
	api := b.ItemPackage.Interface.(*GameInterface.GameInterface)
	// 初始化
	newPackage := *b.ItemPackage
	newPackage.Item.Basic.Name = "writable_book"
	newRequest := DefaultItem{ItemPackage: &newPackage}
	err := newRequest.WriteData()
	if err != nil {
		return fmt.Errorf("OpenBook: %v", err)
	}
	// 获取成书
	err = api.ChangeSelectedHotbarSlot(b.ItemPackage.AdditionalData.HotBarSlot)
	if err != nil {
		return fmt.Errorf("OpenBook: %v", err)
	}
	err = api.ClickAir(b.ItemPackage.AdditionalData.HotBarSlot)
	if err != nil {
		return fmt.Errorf("OpenBook: %v", err)
	}
	// 打开成书
	for key, value := range b.BookData.Pages {
		api.WritePacket(&packet.BookEdit{
			ActionType:    packet.BookActionReplacePage,
			InventorySlot: b.ItemPackage.AdditionalData.HotBarSlot,
			Text:          value,
			PageNumber:    byte(key),
		})
	}
	// 写入文字
	if b.ItemPackage.Item.Basic.Name == "written_book" {
		api.WritePacket(&packet.BookEdit{
			ActionType:    packet.BookActionSign,
			InventorySlot: b.ItemPackage.AdditionalData.HotBarSlot,
			Title:         b.BookData.Title,
			Author:        b.BookData.Author,
		})
	}
	// 签名处理
	err = api.AwaitChangesGeneral()
	if err != nil {
		return fmt.Errorf("OpenBook: %v", err)
	}
	// 等待更改
	if b.ItemPackage.Item.Basic.Name == "written_book" && b.ItemPackage.Item.Basic.Count > 1 {
		uniqueId, err := api.BackupStructure(
			GameInterface.MCStructure{
				BeginX: b.ItemPackage.AdditionalData.Position[0],
				BeginY: b.ItemPackage.AdditionalData.Position[1],
				BeginZ: b.ItemPackage.AdditionalData.Position[2],
				SizeX:  1,
				SizeY:  1,
				SizeZ:  1,
			},
		)
		if err != nil {
			return fmt.Errorf("OpenBook: %v", err)
		}
		defer api.RevertStructure(uniqueId, b.ItemPackage.AdditionalData.Position)
		err = api.CopyItem(
			b.ItemPackage.AdditionalData.HotBarSlot,
			b.ItemPackage.AdditionalData.Position,
			b.ItemPackage.Item.Basic.Count,
		)
		if err != nil {
			return fmt.Errorf("OpenBook: %v", err)
		}
	}
	// 对于堆叠型物品的处理
	return nil
	// 返回值
}
