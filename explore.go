package main

type exploreFunc func() try

/*
Explorer returns an exploreFunc with initial max start and a stride value.

The nominal single case would be Explorer(0, 1), which returns a generator
function that yields the sequence:

    (0, 0)

    (0, 1)
    (1, 0)
    (1, 1)

    (0, 2)
    (1, 2)
    (2, 0)
    (2, 1)
    (2, 2)

    (0, 3)
    ...

Multiple generator functions can be used to divide a sequence by adjusting the
start offset and stride accordingly.  See splitExplore for a convenience
function to help with this.
*/
func Explorer(start, stride int) exploreFunc {
	max := start
	i, j := 0, 0

	return func() try {
		if i >= max && j > max {
			i, j = 0, 0
			max += stride
		}
		if i <= max-1 {
			ri := i
			i++
			return try{ri, max}
		}
		// known: j <= max
		rj := j
		j++
		return try{max, rj}
	}
}

// splitExplore will return a slice of exploreFuncs to split the sequence across
// n generators. These generator functions do not share state and can be safely
// used in parallel.
//
// Example: splitExplore(4) returns []exploreFunc{ Explorer(0,4), Explorer(1,4),
// Explorer(2,4), Explorer(3,4)}.
//
//     f(0, 4)        f(1, 4)         f(2, 4)         f(3, 4)
//     -------        -------         -------         -------
//     {0 0}          {0 1}           {0 2}           {0 3}
//                    {1 0}           {1 2}           {1 3}
//     {0 4}          {1 1}           {2 0}           {2 3}
//     {1 4}                          {2 1}           {3 0}
//     {2 4}          {0 5}           {2 2}           {3 1}
//     {3 4}          {1 5}                           {3 2}
//     {4 0}          {2 5}           {0 6}           {3 3}
//     {4 1}          {3 5}           {1 6}
//     {4 2}          {4 5}           {2 6}           {0 7}
//     {4 3}          {5 0}           {3 6}           {1 7}
//     {4 4}          {5 1}           {4 6}           {2 7}
//                    {5 2}           {5 6}           {3 7}
//     {0 8}          {5 3}           {6 0}           {4 7}
//     {1 8}          {5 4}           {6 1}           {5 7}
//     {2 8}          {5 5}           {6 2}           {6 7}
//     {3 8}                          {6 3}           {7 0}
//     {4 8}          {0 9}           {6 4}           {7 1}
//     {5 8}          {1 9}           {6 5}           {7 2}
//     {6 8}          {2 9}           {6 6}           {7 3}
//     {7 8}          {3 9}                           {7 4}
//     {8 0}          {4 9}           {0 10}          {7 5}
//     {8 1}          {5 9}           {1 10}          {7 6}
//     ...            ...             ...             ...
func splitExplore(n int) []exploreFunc {
	if n <= 0 {
		return []exploreFunc{}
	}
	res := make([]exploreFunc, n)
	for i := 0; i < n; i++ {
		res[i] = Explorer(i, n)
	}
	return res
}
