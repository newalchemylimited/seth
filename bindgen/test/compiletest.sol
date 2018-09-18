pragma experimental ABIEncoderV2;
pragma solidity >= 0.4.23;

contract Test {
   // addr, err := sender.Create(TestCode, nil, "(uint16,string)", uint16(123), "hi how are you")

    uint16 public cuint16;
    string public cstring;

    constructor(uint16 cuint16val, string cstringval) public {
        cuint16 = cuint16val;
        cstring = cstringval;
    }

    // structs
    Person[] public people;

    struct Person {
        string name;
        uint8 age;
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

    // event
    event SomethingHappened(uint16 uint16Val, address addressVal, string stringVal, bytes bytesVal);
    function sendTestEvent(uint16 uint16Val, string stringVal, bytes bytesVal) public {
        emit SomethingHappened(uint16Val, msg.sender, stringVal, bytesVal);
    }

    // bytes32
    bytes32 public bytes32Val;

    function setBytes32Val(bytes32 val) public {
        bytes32Val = val;
    }

    // bytes
    bytes public bytesVal;

    function setBytesVal(bytes val) public {
        bytesVal = val;
    }

    // string

    string public stringVal;

    function setStringVal(string val) public {
        stringVal = val;
    }

    function must_throw() public {
        require(false, "should throw");
    }

    function double_this(uint32 To_be_Doubled) public pure returns(uint64 doubled) {  
        return To_be_Doubled * 2;
    }

}
