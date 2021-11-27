import time
global glob
localv=5
def testFuncAAA(a):
	list=["a","b","c"]
	list.append(4)
	val=list[3]//5
	list.append((1,2,4))
	val=val|list[4][2]
	out=((((val<<8)^5)>>5)&80)
	out+=time.time()
	o=-out
	out+=(o/500)
	for i in range(1,10):
		list.append(i)
	alist=list[4]+(1)
	print(alist)
	del alist
	return out
