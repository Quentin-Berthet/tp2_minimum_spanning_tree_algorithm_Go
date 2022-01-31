# TP2 - Minimum Spanning Tree Algorithm

## GHS Algorithm

- Spanning Tree Fragment
    - Any (connected) sub-tree of a MST
- Minimum weight outgoing edge
    - An edge adjacent to the fragment with the smallest weight and that does not create a cycle
- MST Property I
    - Given a fragment of an MST, let _e_ be a minimum-weight outgoing edge of the fragment.
    - Then joining _e_ and its adjacent non-fragment node to the fragment yields another fragment of an MST
- MST Property II
    - If all the edges of a connected graph have different weights, then the MST is unique
- MST based on properties 1 & 2
    - Enlarge fragments in any order (property 1)
    - Combine fragments with a common node
- Asynchronous GHS
    - Every fragment F has a level L(F) >= 0
        - At the beginning, each node is a fragment of level 0
    - Two types of merges : absorption and join
    - At the beginning, nodes know only the weights of their neighbours
    - At the end, every node knows which of its links belong to the MST and which do not
    - Levels
        - Every fragment has an associated level that has an impact on combining fragments
        - A fragment with a single node is defined to be at level 0
        - The combination of two fragments depend on the levels of fragments
            - F want to connect to F', L < L', L'
            - F want to connect to F', L = L', new fragment F'', L'' = L + 1
                - The identity of a fragment is the weight of its core
                    - Core : combining edge 
    - State
        - Each node has a state
            - Sleeping, initial state
            - Find, during fragment's search for a minimal outgoing edge
            - Found otherwise, when a minimal outgoing edge was found
    - Description
        - Each non-zero level fragment has an ID, which is the id of th core edge in the fragment
        - During exec. each node can classify each of its edges into 3 categories
            - Branch edges, already been determined to be part of the MST
            - Rejected edges, already been determined NOT to be part of the MST
            - Basic edges, otherwise
        - For level 0 fragments, each awakened node :
            - Choose its min-weight edge and marks that edge as a branch edge
            - Send a message via the branch edge to notify the node
            - Wait for a message from the other end of the edge
            - The edge chosen by both nodes it connects becomes the core with level 1
        - For non-zero level fragment, 3 stages
            - Broadcast, the two nodes adjacent to the core broadcast messages to the rest of the nodes in the fragment via the BRANCH EDGE. Each message contains, the ID and level of the fragment
            - Convergecast, leaf node send the min weight, non-leaf edge wait for n-1 messages, and pick the min weight
            - Change core, the two nodes connected by the core can inform each other of the best edges they received. Then, they can identify the minimum outgoing edge from the entire fragment. A message will be sent FROM the CORE to the min edge via a path of branch edges. Finally, a message will be sent out via the chosen outgoing edge to request to combine the two fragments.
            
[Source](https://www.youtube.com/watch?v=9wR-XRju5NM)
## Construire le projet
* Exécutez `make`

