# Golang Katas

I wanted to do some coding katas to help me get better at algorithms. Since AI will steal my job one day, I thought I'd get it to help me prolong my employment by creating the tool.

I didn't vibe code it, I wanted it to use Golang compiled to WASM.

I decided to use [Yaegi](https://github.com/traefik/yaegi) so I can compile Go code and then got it to generate the code accordingly. I had a look at it and manually made a few changes but on the whole it worked quite well.

## Development

Ideally you would be [NixOS](https://nixos.org/) or have the [Nix pacakage manager](https://wiki.nixos.org/wiki/Nix_\(package_manager\)) and then you can just run:

```sh
nix develop
```

This will install everything you need.

To build the tool locally:

```sh
make clean single # for single file output
make clean split # for multi file output
```

Then you will find the generated HTML file at `dist/index.html`.

## Katas

The list of katas is:

* Add Two Numbers _This is for demo purposes_
* Two Sum II (Sorted, Two Pointers)
* Container With Most Water
* 3Sum
* Sort Colors (Dutch National Flag)
* Linked List Cycle Detection (Floyd's Tortoise & Hare)
* Find the Duplicate Number (Floyd's Cycle Method)
* Longest Substring Without Repeating Characters
* Minimum Window Substring
* Merge Intervals
* Insert Interval
* Missing Number (0..n)
* Find All Duplicates in an Array
* Reverse Linked List (slice representation)
* Reverse Nodes in k-Group (slice representation)
* Valid Parentheses
* Min Stack (operation simulation)
* Next Greater Element I
* Largest Rectangle in Histogram
* Two Sum (Return Indices)
* Group Anagrams
* Binary Tree Level Order Traversal (array input)
* Minimum Depth of Binary Tree (array input)
* Path Sum (root-to-leaf equals target)
* Connected Components in an Undirected Graph
* Number of Islands
* Max Area of Island
* Online Medians (Two Heaps)
* Sliding Window Median (Two Heaps)
* Top K Frequent Elements
* Merge k Sorted Lists (as arrays)
* Smallest Range Covering Elements from k Lists
* Subsets (Power Set)
* Combination Sum
* N-Queens (Count Solutions)
* Word Search (exist)
* Search in Rotated Sorted Array
* Find Peak Element
* Single Number (XOR)
* Maximum XOR of Two Numbers in an Array
* Interval Scheduling (Max Non-Overlapping)
* Jump Game II (Min Jumps)
* Trie (Insert/Search/StartsWith Simulation)
* Word Search II (find words)
* Course Schedule II (Topo Order)
* Alien Dictionary (Topo of Characters)
* Graph Valid Tree (Union-Find)
* Redundant Connection (Union-Find)
* Longest Consecutive Sequence
* Contains Duplicate III (Ordered Set)
* Subarray Sum Equals K (Count)
* Maximum Size Subarray Sum Equals K

You can add or remove katas by modifying `katas.json` and rebuilding the application.
