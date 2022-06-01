package upgrade

func Upgrade() {
	storageRoot := Policy_1()
	Policy_2(storageRoot)
	Policy_3(storageRoot)
	Policy_4(storageRoot)
	Policy_5(storageRoot)
	Policy_6(storageRoot)
	Policy_7(storageRoot)
}
