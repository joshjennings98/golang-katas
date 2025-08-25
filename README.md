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

* Two Sum II (Sorted, Two Pointers)
* Container With Most Water
* 3Sum
* Sort Colors (Dutch National Flag)
* Linked List Cycle Detection (Floyd's Tortoise & Hare)
* Longest Substring Without Repeating Characters
* Minimum Window Substring
* Merge Intervals
* Insert Interval
* Valid Parentheses
* Min Stack (operation simulation)
* Next Greater Element I
* Largest Rectangle in Histogram
* Group Anagrams
* Binary Tree Level Order Traversal (array input)
* Path Sum (root-to-leaf equals target)
* Connected Components in an Undirected Graph
* Number of Islands
* Online Medians (Two Heaps)
* Sliding Window Median (Two Heaps)
* Top K Frequent Elements
* Subsets (Power Set)
* Combination Sum
* Word Search (exist)
* Search in Rotated Sorted Array
* Find Peak Element
* Interval Scheduling (Max Non-Overlapping)
* Jump Game II (Min Jumps)
* Trie (Insert/Search/StartsWith Simulation)
* Course Schedule II (Topo Order)
* Alien Dictionary (Topo of Characters)
* Graph Valid Tree (Union-Find)
* Redundant Connection (Union-Find)
* Longest Consecutive Sequence
* Contains Duplicate III (Ordered Set)
* Subarray Sum Equals K (Count)
* Maximum Size Subarray Sum Equals K

You can add or remove katas by modifying `katas.json` and rebuilding the application.
