function circle(radius) {
	let arr=[];
	let radius_square=Math.pow(radius,2);
	for(let i=0;i<=radius;i++) { // x
		let first_quadrant_val=Math.sqrt(radius_square-Math.pow(i,2));
		arr.push([i,first_quadrant_val]);
		if(first_quadrant_val)arr.push([i,-first_quadrant_val]);
		if(i)arr.push([-i,first_quadrant_val]);
		if(first_quadrant_val&&i)arr.push([-i,-first_quadrant_val]);
	}
	return arr;
}
console.log(circle(10));