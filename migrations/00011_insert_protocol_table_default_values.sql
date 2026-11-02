-- +goose Up
INSERT INTO protocols (name, description) VALUES
    ('source', 'Minecraft, CS2, CS:GO, ARK: Survival Evolved, Palworld, Project Zomboid, Garry''s Mod, Team Fortress 2, Conan Exiles, Factorio, Rust (Legacy)'),
    ('websocket', 'Rust (WebRCON), V Rising, Eco'),
    ('telnet', '7 Days to Die, Empyrion: Galactic Survival, Battlefield (BF3, BF4, BFBC2), Terraria (TShock)');

-- +goose Down
DELETE FROM protocols WHERE name IN ('source', 'websocket', 'telnet');