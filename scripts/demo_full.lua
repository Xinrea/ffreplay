-- demo_full.lua
-- Full feature demo: party list, telegraphs, buffs, text, walk animation, head markers

local ff = require("ff")

print("=== FFReplay Full Feature Demo ===")

-- 1. Setup
ff.load_map(77)
ff.sleep(300)

-- 2. Create boss
local boss = ff.create_boss("Omega", 9999, 0, 0, 10)
print("Boss created")

-- 3. Create players (party list should update dynamically)
local mt = ff.create_player("DarkKnight", 0, -10)
local st = ff.create_player("Gunbreaker", 3, -8)
local h1 = ff.create_player("WhiteMage", 6, 0)
local h2 = ff.create_player("Scholar", -6, 0)
local d1 = ff.create_player("Dragoon", -3, -10)
local d2 = ff.create_player("Ninja", 3, -10)
local d3 = ff.create_player("Bard", -8, 4)
local d4 = ff.create_player("BlackMage", 8, 4)

print("Party of 8 created")
ff.sleep(500)

-- 4. Place waymarks
ff.add_waymark(0, 0, -15)   -- A north
ff.add_waymark(1, 0, 15)    -- B south
ff.add_waymark(2, -15, 0)   -- C west
ff.add_waymark(3, 15, 0)    -- D east
print("Waymarks placed")

-- 5. Demo: AoE telegraph (circle)
print("Drawing AoE circle telegraph...")
ff.draw_circle(0, 0, 8, 3000)  -- 8m radius, 3s duration
ff.sleep(1000)

-- 6. Demo: Rectangular telegraph (cleave)
print("Drawing cleave telegraph...")
ff.draw_rect(0, -5, 4, 10, 2000)  -- 4m x 10m, 2s duration
ff.sleep(500)

-- 7. Demo: Text annotation
print("Drawing text annotation...")
ff.draw_text(0, -12, "Boss AoE - Move out!", 3000)
ff.sleep(500)

-- 8. Demo: Head markers on players
print("Setting head markers...")
mt:set_headmarker(1)
st:set_headmarker(2)
h1:set_headmarker(3)
h2:set_headmarker(4)
ff.sleep(1000)

-- 9. Demo: Walk animation
print("Tank walking to boss...")
mt:walk(0, -2, 1500)  -- walk to (0, -2) over 1.5s
ff.sleep(2000)

print("Boss walking to side...")
boss:walk(-5, 0, 2000)
ff.sleep(2500)

-- 10. Demo: Apply buff (using non-existent IDs for visual demo only)
print("Applying buffs...")
mt:apply_buff(1, 10000, 2)   -- buff id 1, 10s, 2 stacks
h1:apply_buff(2, 15000, 1)   -- buff id 2, 15s, 1 stack
ff.sleep(1000)

-- 11. Demo: Remove buff and head marker
print("Clearing head markers...")
st:clear_headmarker()
h2:clear_headmarker()
ff.sleep(500)

print("Removing buff...")
mt:remove_buff(1)

-- 12. Camera overview
print("Camera sweep...")
ff.camera_set_pos(0, 0)
ff.camera_zoom(-5)
ff.sleep(1000)

-- Boss returns center
boss:walk(0, 0, 1500)
ff.sleep(2000)

print("=== Full Demo Complete! ===")
