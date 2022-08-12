package guild

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/omega/defines"
	"strconv"
)

// op权限菜单
func (b *Guild) setOpMenu(name string) {
	go func() {
		//查看是否为有权的人
		n := <-b.GetPlayerName(b.TargetOfSetGuildLb)
		fmt.Println("[提示信息] 在线符合修改公会权限的人员:", n)
		if b.FindmapKey(n, name) {
			b.sendMenu(name, b.KeyTitle["op隐藏菜单"])
			//b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.KeyTitle["op隐藏菜单"])
			b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
				if len(chat.Msg) > 0 {
					switch chat.Msg[0] {
					case "0":
						b.findGuildData(chat.Name)
					case "1":
						b.chargeGuildData(chat.Name)
					}

				}

				return true
			})
		} else {
			b.sayto(name, "[错误] 你没有对应权限")
		}

	}()

}

// 商城菜单
func (b *Guild) StarGuild(name string) {
	//判断对方公会是否达到对应级别
	if k, v, _ := b.CheckInGuild(name); v {
		if b.GuildData[k].Power >= b.ThePermissionsOfGuild["商店功能"] {
			//达到级别则弹出商店
			msg := "§b[公会商店]\n输入对应数字来购买对应商品\n"
			for i, j := range b.StarGuilds {
				msg = msg + b.FormateMsg(b.KeyTitle["公会菜单商品显示"], "商品", j.name)
				msg = b.FormateMsg(msg, "价格", j.Price)
				msg = b.FormateMsg(msg, "i", i)
				msg = msg + "\n"
			}
			b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
			b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
				for _i, _j := range b.StarGuilds {
					i, j := _i, _j
					//fmt.Println(i, "i", "\nmsg:", chat.Msg[0])
					if len(chat.Msg) > 0 && i == chat.Msg[0] {
						//fmt.Println("触发")
						cmd := b.FormateMsg(j.CheckCmds, "购买者", chat.Name)
						cmd = b.FormateMsg(cmd, "计分板", j.Score)
						cmd = b.FormateMsg(cmd, "购买价格", j.Price)
						fmt.Println("cmd:", cmd)
						//防止错位而专门设置的新名字
						newName := name
						b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(cmd, func(output *packet.CommandOutput) {
							if output.SuccessCount > 0 {

								for _, c := range j.Cmds {
									c = b.FormateMsg(c, "购买者", newName)
									c = b.FormateMsg(c, "计分板", j.Score)
									c = b.FormateMsg(c, "购买价格", j.Price)
									//发送指令并查看是否正确
									fmt.Println("c:", c)
									b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(c, func(output *packet.CommandOutput) {
										fmt.Println("ouputmsg:", output.OutputMessages)
									})
								}
							} else {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", newName), b.KeyTitle["商店购买失败提示"])
							}
						})
					}
				}

				return true
			})
		}
	} else {
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.KeyTitle["权限不足时提示"])
	}
}

// 购买公会菜单
func (b *Guild) BuyGuild(name string) {
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(b.FormateMsg(b.FormateMsg(b.FormateMsg(fmt.Sprintf("testfor %v", b.TartgetBuy), "player", name), "购买计分板", b.DictScore["购买公会计分板"]), "价格", b.Price), func(output *packet.CommandOutput) {
		fmt.Println("outputmessage : ", output.OutputMessages)
		if output.SuccessCount > 0 {
			b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§a[输入公会名字]:"))
			if b.Frame.GetGameControl().SetOnParamMsg(name, func(NewChat *defines.GameChat) (catch bool) {
				GuildName := NewChat.Msg[0]
				//检查名字重合与坐标重合与否（）
				if b.ChecklGuildName(GuildName) == false { //&& b.CheckIsoverlap() {
					b.WriteGuildData(name, GuildName)

				} else {
					b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§c[名字已经被注册了] 请重新购买"))
				}

				return true
			}) == nil {

			}

		} else {
			b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("§c§l[错误]§a 无法购买请满足选择器:§b\n%v\n", b.FormateMsg(b.FormateMsg(b.FormateMsg(fmt.Sprintf("%v", b.TartgetBuy), "player", name), "购买计分板", b.DictScore["购买公会计分板"]), "价格", b.Price)))
		}
	})
}

// 设置传送点菜单
func (b *Guild) setTpPos(name string) {
	func() {
		guildname, _, powernum := b.CheckInGuild(name)
		if powernum == 4 {
			pos := <-b.Frame.GetGameControl().GetPlayerKit(name).GetPos("@a[name=[player]]")
			if b.GuildData[guildname].IsTerr {
				if b.CheckInGuildPlace(guildname, pos) {
					b.GuildData[guildname].CenterPos = pos
					b.sayto(name, b.KeyTitle["传送坐标设置成功提示词"])
				} else {
					b.sayto(name, b.KeyTitle["有地皮时未站在地皮范围内设置传送点提示词"])
				}
			} else {

				b.GuildData[guildname].CenterPos = pos
				b.sayto(name, b.KeyTitle["传送坐标设置成功提示词"])
			}

		} else {
			b.sayto(name, b.KeyTitle["权限不足时提示"])
		}

	}()
}

// 提升公会权限菜单
func (b *Guild) UpgradeGuild(name string) {
	guildname, _, powernum := b.CheckInGuild(name)
	if powernum == 4 {
		b.sayto(name, b.FormateMsg(b.KeyTitle["提示公会当前权限"], "公会权限", strconv.Itoa(b.GuildData[guildname].Power)))
		if b.GuildData[guildname].Power < 4 {
			go func() {
				list := <-b.GetPlayerName(fmt.Sprintf("@a[name=\"%v\",scores={%v=%v..}]", name, b.DictScore["贡献计分板"], b.UpgradePrice[strconv.Itoa(b.GuildData[guildname].Power)]))
				if b.FindmapKey(list, name) {
					b.CmdSender(fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", name, b.DictScore["贡献计分板"], b.UpgradePrice[strconv.Itoa(b.GuildData[guildname].Power)]))
					b.GuildData[guildname].Power++
					b.sayto(name, b.KeyTitle["提升公会权限成功"])
				} else {
					b.sayto(name, b.KeyTitle["提升工会权限失败提示词"])
				}
			}()
		} else {
			b.sayto(name, b.KeyTitle["公会权限已达到最高"])
		}

	} else {
		b.sayto(name, b.KeyTitle["权限不足时提示"])
	}

}

// 会长菜单&&管理员菜单
func (b *Guild) MasterMenu(name string) {
	msg := b.KeyTitle["管理菜单显示"]
	//fmt.Print(msg)
	//n,ok,PowerNum :=b.CheckInGuild(name)

	//必须是管理员级别才能菜单执行 否则不予理会
	if n, _, PowerNum := b.CheckInGuild(name); PowerNum >= 2 {
		//发送菜单等待回执
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				//fmt.Print("--test--320")

				go func() {
					k, isOK := <-b.GetPlayerName("@a")
					if isOK {
						//fmt.Print("--test--320")
						switch chat.Msg[0] {
						case "0":

							//fmt.Print("name:", k)
							msgOfPlayer := "§e[请输入对应数字邀请对方]\n"

							for i, PlayerName := range k {
								msgOfPlayer = msgOfPlayer + fmt.Sprintf("[%v] %v", i, PlayerName)
							}
							b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msgOfPlayer)
							b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
								if len(Newchat.Msg) > 0 {
									if _, isok := k[Newchat.Msg[0]]; isok {

										if yes, _ := b.CheckIsMaster(k[Newchat.Msg[0]]); yes {

											b.sayto(name, b.KeyTitle["邀请对方加入公会失败提示词"])
										} else {
											b.Invite(n, k[Newchat.Msg[0]], k)
										}

									}
								} else {
									b.FromatePrint("输入错误", "提示信息")
								}

								return true
							})

						case "1":
							//达到权限才能使用否则警告
							if PowerNum >= b.ThePermissionsOfMember["可以改变他人权限"] {
								MenuMsg := "§b[设置他人权限]"
								ListMap := make(map[string]string)
								num := 0
								for i, _ := range b.GuildData[n].Member {
									MenuMsg = MenuMsg + fmt.Sprintf("[%v] %v", num, i)
									ListMap[strconv.Itoa(num)] = i
									num++
								}
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), MenuMsg)
								b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
									if len(Newchat.Msg) > 0 {
										if _, isok := ListMap[Newchat.Msg[0]]; isok {
											b.SetOtherPower(Newchat.Name, ListMap[Newchat.Msg[0]], n)
										} else {
											b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", Newchat.Name), "[请输入有效数字]")
										}
									}

									return true
								})
							} else {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.KeyTitle["权限不足时提示"])
							}
						case "2":
							/*
								if PowerNum >= b.ThePermissionsOfMember["可以发布公告"] {
									b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), "[请输入公告内容]")
									b.Frame.GetGameControl().SetOnParamMsg(chat.Name, func(Newchats *defines.GameChat) (catch bool) {
										b.GuildData[n].announcement = append(b.GuildData[n].announcement, Newchats.Msg[0])
										b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", Newchats.Name), fmt.Sprintf("[发布公告] 内容为:%v", Newchats.Msg[0]))
										return true
									})
								} else {
									b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.KeyTitle["权限不足时提示"])
								}*/
							b.setTpPos(name)
						case "3":

							if b.GuildData[n].Power < b.ThePermissionsOfGuild["设置二级保护区域"] {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.KeyTitle["权限不足时提示"])
							} else if PowerNum >= b.ThePermissionsOfMember["可以设置二级保护区"] {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), "[请站在起点] 输入任意文字")
								fmt.Println("公会名字为:", n)
								b.setSpPlace(chat.Name, n)

							} else {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", chat.Name), b.KeyTitle["权限不足时提示"])

							}
						case "4":
							b.kickMember(n, chat.Name)
						case "5":
							delete(b.GuildData, n)
							b.sayto(name, "[删除成功]")
						case "6":
							b.getGuildDataD(name, n, *b.GuildData[n])
						case "7":
							b.DeleteGuildMember(name, *b.GuildData[n])
						case "8":
							b.UpgradeGuild(name)
						}

					} else {
						b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), "[错误] 没有找到目标")
					}
				}()

			}

			return true
		})
	}

}

// op 权限菜单子菜单 ----改变公会的数据
func (b *Guild) chargeGuildData(name string) {
	b.sayto(name, "[输入想要修改的公会名字]")
	theGuildMap := b.getGuildMap()
	b.formateGuildNameMenu(name, b.KeyTitle["设置公会权限菜单模板"], theGuildMap)
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 {
			if guildname, ok := theGuildMap[chat.Msg[0]]; ok {
				b.sayto(name, "[输入想要修改的权限]\n0 :删除领地\n1 :修改领地权限")
				b.Frame.GetGameControl().SetOnParamMsg(name, func(newchat *defines.GameChat) (catch bool) {
					if len(newchat.Msg) > 0 {
						switch newchat.Msg[0] {
						case "0":
							delete(b.GuildData, guildname)
							b.sayto(name, "[删除成功]")
						case "1":
							b.sayto(name, "[输入权限等级]")
							b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
								if len(Newchat.Msg) > 0 {
									if newchat.Msg[0] == "1" || newchat.Msg[0] == "2" || newchat.Msg[0] == "3" || newchat.Msg[0] == "4" {
										b.GuildData[guildname].Power, _ = strconv.Atoi(Newchat.Msg[0])
										b.sayto(name, "[修改成功]")
									} else {
										b.sayto(name, "【输入有效数字】")
									}
								}
								return true
							})
						}

					}
					return true
				})
			} else {
				b.sayto(name, "[请输入有效数字]")
			}
		}
		return true
	})
}

// 设置公会权限菜单
// 暂时废弃
func (b *Guild) setGuildPowerMenu(name string) {
	list := make(map[string]string)
	msg := "[输入对应的数字进行操纵]"
	if len(b.GuildData) > 0 {
		//获取公会名字并成列表给用户并让它选择

		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), msg)
		b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
			if len(chat.Msg) > 0 {
				//k为公会名字
				if k, ok := list[chat.Msg[0]]; ok {
					b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("[提示] 你选择的是%v公会 请输入更改权限大小(1/2/3/4)", k))
					b.Frame.GetGameControl().SetOnParamMsg(name, func(Newchat *defines.GameChat) (catch bool) {
						if len(Newchat.Msg) > 0 {

							if Newchat.Msg[0] == "1" || Newchat.Msg[0] == "2" || Newchat.Msg[0] == "3" || Newchat.Msg[0] == "4" {
								num, _ := strconv.Atoi(Newchat.Msg[0])
								b.GuildData[k].Power = num
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("[修改成功] %v 更改权限为%v", k, strconv.Itoa(b.GuildData[k].Power)))

							} else {
								b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), fmt.Sprintf("请输入有效数字"))

							}

						}
						return true
					})
				}

			}
			return true
		})

	} else {
		b.Frame.GetGameControl().SayTo(fmt.Sprintf("@a[name=\"%v\"]", name), b.KeyTitle["暂无公会存在时提示词"])
	}
}

// 发送命令如果失败则回执失败消息
func (b *Guild) CmdSender(str string) {
	b.Frame.GetGameControl().SendCmdAndInvokeOnResponse(str, func(output *packet.CommandOutput) {
		if output.SuccessCount > 0 {

		} else {
			fmt.Println("[提示信息 执行指令失败] 错误信息如下:", output.OutputMessages, "\n执行的指令为:", str)
		}
	})
}

// 金钱交易菜单
func (b *Guild) bourseMenu(name string) {
	b.sayto(name, b.KeyTitle["兑换贡献时提示词1"])
	b.Frame.GetGameControl().SetOnParamMsg(name, func(chat *defines.GameChat) (catch bool) {
		if len(chat.Msg) > 0 && b.CheckIsNum(chat.Msg[0]) {

			go func() {
				list := <-b.GetPlayerName(fmt.Sprintf("@a[name=\"%v\",scores={%v=%v..}]", name, b.DictScore["兑换贡献计分板"], chat.Msg[0]))
				if b.FindmapKey(list, name) {
					b.CmdSender(fmt.Sprintf("scoreboard players remove @a[name=\"%v\"] %v %v", name, b.DictScore["兑换贡献计分板"], chat.Msg[0]))
					b.CmdSender(fmt.Sprintf("scoreboard players add @a[name=\"%v\"] %v %v", name, b.DictScore["贡献计分板"], chat.Msg[0]))
					b.sayto(name, b.KeyTitle["兑换贡献值成功提示词"])
				} else {
					b.sayto(name, b.KeyTitle["兑换贡献值计分板积分不足提示词"])
				}
			}()

		} else {
			b.sayto(name, "[请输入正确的数字]")
		}

		return true
	})
}
