# Event bus

## 목표
WebSocket만 사용해서 구현되는 이벤트 전달 구조

## 용어
* Primary Peer: 실행 시 Argument로 지정된 Server
* Network ID: 같은 Network 소속이라는 것을 알 수 있는 일종의 토큰
* Terminal Node: Peer가 붙을 수 없는 Node
* Root Node: Primary Peer가 없는 Node

## Note

* 기본적으로는 Peer에서 발생된 이벤트를 다른 Peer로 전달
* 각 Node는 일정 주기마다 자신의 정보를 알리는 Event 발생
    * 자신의 ConnectionBack Score(비율)
    * 자신의 Peer connection List
* Node List 는 각 Node에서 따로 관리
    * Peer List
        * 실제로 connection을 맺고 있는 List
    * Node List
        * 전체 Network 상의 List
* Node 마다 Connection Check 시 그 주소로 접근 해보는 ConnectionBack 과정을 하여 자신의 공인 IP 및 주소를 확인
* Primary Peer는 주기적으로 재연결
* 각 Node는 Peer의 갯수를 Argument로 받고 해당 갯수보다 Peer 갯수가 적으면 Peer 연결을 시도
    * 다른 node 에서 자신의 정보를 알리는 Event를 보고 그중 하나를 선택
    * 각 node는 서로 다른 connection 전략을 취할 수 있음
* Network Id가 다른 접근은 모두 무시함
* Terminal Node 는 접속을 하되 자신을 알리는 Event를 발생 시키지 않음.
* 각 Message는 Expire 시간과 ID 가 있음
    * 모든 Node는 Expire시간 까지는 ID를 보존 해야 함
    * 같은 ID로 두 번 전달 되는 이벤트는 무시
    * Target이 연결되면 Expire 되지 않은 모든 Event를 준다.
        * 만약 Event 처리가 절대 두 번 이상 안되야 하면 어느 이벤트를 처리했는지 Persistance 하게 관리할 책임은 각 Node에 있음.
* Storage 는 쓰지 않고 오직 in-memory-db 만 사용해서 구현
    * 데이터는 휘발성이다.
    * 단 Terminal Node는 구현에 따라 Persistancy가 필요할수 있다.
* advertise address를 직접 지정 하지 않고 쓸수 있으면 좋겠다.
