pragma experimental ABIEncoderV2;
pragma solidity >= 0.4.23;

contract Test {
    uint32 public counter;

    string public name;
    

    
    Person[] public people;

    struct Person {
        string name;
        uint8 age;
    }

    constructor() public {
    }
    
    function addElliot() public {
        people.push(Person({
            name: "elliot",
            age: 34
        }));
    }

    function addPerson(Person p) public {
        people.push(p);
    }
       
    function allPeople() public view returns(Person[] everyone) {
        return people;
    }


    // bytes32
    bytes32 public bytes32val;

    function setBytes32val(bytes32 val) public {
        bytes32val = val;
    }

    // bytes
    bytes public bytesval;

    function setBytesval(bytes val) public {
        bytesval = val;
    }

    // string
    string public stringval;

    function setStringval(string val) public {
        stringval = val;
    }

    function value() public view returns(uint32 current_value) {
        return counter;
    }

    function must_throw() public {
        require(false, "should throw");
    }

    function double_this(uint32 To_be_Doubled) public pure returns(uint64 doubled) {  
        return To_be_Doubled * 2;
    }

    function inc() public {
        counter = counter + 1;
    }

    function incBy(uint32 i) public {
        counter = counter + i;
    }

    function SetName(string newName) public {
        name = newName;
    }

    function getBig() public view returns (uint mr_big) {
        return counter;
    }
}
