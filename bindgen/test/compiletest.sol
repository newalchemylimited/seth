pragma solidity >=0.4.10;

contract Test {
	uint public counter;

	function Test() {
	}

	function value() constant returns(uint) {
		return counter;
	}

	function mustThrow() {
		require(false);
	}

	function inc() {
		counter = counter + 1;
	}
}
