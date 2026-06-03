-- demo_boss.lua
-- FF14 raid guide demo: Boss & mechanic setup
-- Demonstrates: boss creation, waymarks, multi-player positioning

local ff = require("ff")

print("=== FFReplay Script Demo: Boss & Mechanics ===")

-- 1. Setup map
ff.load_map(77)
print("Map loaded: The Omega Protocol")

ff.sleep(500)

-- 2. Create a boss at the center
print("Creating boss...")
local boss = ff.create_boss("Omega", 1000, 0, 0, 8)
print("Boss created with ID: " .. boss:id())

ff.sleep(500)

-- 3. Create full party
print("Creating party...")
local mt = ff.create_player("DarkKnight", -3, -8)
local st = ff.create_player("Gunbreaker", 3, -8)
local h1 = ff.create_player("WhiteMage", 5, 2)
local h2 = ff.create_player("Astrologian", -5, 2)
local d1 = ff.create_player("Monk", -3, -10)
local d2 = ff.create_player("Ninja", 3, -10)
local d3 = ff.create_player("Bard", -8, 5)
local d4 = ff.create_player("Summoner", 8, 5)

ff.sleep(500)

-- 4. Place waymarks around the arena
print("Placing waymarks...")
-- A B C D markers in cardinal positions
ff.add_waymark(0, 0, -15)    -- A (north)
ff.add_waymark(1, 0, 15)     -- B (south)
ff.add_waymark(2, -15, 0)    -- C (west)
ff.add_waymark(3, 15, 0)     -- D (east)

ff.sleep(500)

-- 5. Move boss to show mechanic
print("Boss moving...")
boss:set_pos(-5, 0)
ff.sleep(300)

print("Boss casting cleave... (facing south)")
boss:face(1.57) -- face south
ff.sleep(500)

print("Boss returning to center...")
boss:set_pos(0, 0)
boss:face(0) -- face north

ff.sleep(500)

-- 6. Camera work
print("Camera zoom out for full arena view...")
ff.camera_zoom(-10)

ff.sleep(500)

print("=== Boss demo complete! ===")
