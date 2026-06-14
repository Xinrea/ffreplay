-- demo_basic.lua
-- Basic FF14 playground script demo
-- Demonstrates: map loading, player creation, positioning, camera control

local ff = require("ff")

print("=== FFReplay Script Demo: Basic ===")

-- 1. Load a map (map ID 77 = The Omega Protocol)
ff.load_map(77)
print("Map loaded")

-- Wait a moment
ff.sleep(500)

-- 2. Create players
print("Creating players...")
local tank = ff.create_player("Warrior", 0, -5)
local healer = ff.create_player("Scholar", 5, 3)
local dps1 = ff.create_player("Dragoon", -5, 3)
local dps2 = ff.create_player("BlackMage", 5, -3)

print("Created " .. tank:id() .. " (Warrior)")
print("Created " .. healer:id() .. " (Scholar)")

ff.sleep(500)

-- 3. Position players in a diamond formation
print("Positioning players...")
tank:set_pos(0, -8)
healer:set_pos(0, 5)
dps1:set_pos(-8, 0)
dps2:set_pos(8, 0)

ff.sleep(1000)

-- 4. Reposition tank
print("Moving tank to center...")
tank:set_pos(0, 0)

-- 5. Camera control
print("Panning camera...")
ff.camera_set_pos(0, 0)

ff.sleep(500)

print("=== Demo complete! ===")
