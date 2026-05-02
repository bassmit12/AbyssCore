-- ─── Warrior Cards ───────────────────────────────────────────────────────────
INSERT INTO deck.card_definitions (name, class, type, cost, effect, rarity, description) VALUES
  ('Strike',        'warrior', 'attack', 1, '{"damage":6}',                                         'common',   'Deal 6 damage.'),
  ('Defend',        'warrior', 'skill',  1, '{"block":5}',                                          'common',   'Gain 5 Block.'),
  ('Bash',          'warrior', 'attack', 2, '{"damage":8,"apply_vulnerable":2}',                    'common',   'Deal 8 damage. Apply 2 Vulnerable.'),
  ('Headbutt',      'warrior', 'attack', 1, '{"damage":9,"put_discard_on_draw":1}',                 'common',   'Deal 9 damage. Put the top card of your discard pile on top of your draw pile.'),
  ('Iron Wave',     'warrior', 'attack', 1, '{"damage":5,"block":5}',                               'common',   'Gain 5 Block. Deal 5 damage.'),
  ('Cleave',        'warrior', 'attack', 1, '{"damage":4,"all_enemies":true}',                      'common',   'Deal 4 damage to ALL enemies.'),
  ('Whirlwind',     'warrior', 'attack', 0, '{"damage_per_energy":5,"all_enemies":true,"x_cost":true}', 'uncommon','Deal 5 damage to ALL enemies X times.'),
  ('Bodyslam',      'warrior', 'attack', 1, '{"damage_equals_block":true}',                         'uncommon', 'Deal damage equal to your current Block.'),
  ('Inflame',       'warrior', 'power',  1, '{"strength":2}',                                       'uncommon', 'Gain 2 Strength.'),
  ('Flex',          'warrior', 'skill',  0, '{"strength_this_turn":2}',                             'common',   'Gain 2 Strength. At end of turn lose 2 Strength.'),

-- ─── Rogue Cards ─────────────────────────────────────────────────────────────
  ('Slice',         'rogue',   'attack', 1, '{"damage":6}',                                         'common',   'Deal 6 damage.'),
  ('Survivor',      'rogue',   'skill',  1, '{"block":8,"discard":1}',                              'common',   'Gain 8 Block. Discard 1 card.'),
  ('Neutralize',    'rogue',   'attack', 0, '{"damage":3,"apply_weak":1}',                          'common',   'Deal 3 damage. Apply 1 Weak.'),
  ('Acrobatics',    'rogue',   'skill',  1, '{"draw":3,"discard":1}',                               'common',   'Draw 3 cards. Discard 1 card.'),
  ('Backflip',      'rogue',   'skill',  1, '{"block":5,"draw":2}',                                 'common',   'Gain 5 Block. Draw 2 cards.'),
  ('Dodge And Roll','rogue',   'skill',  1, '{"block":4,"block_next_turn":4}',                      'common',   'Gain 4 Block. Next turn gain 4 Block.'),
  ('Predator',      'rogue',   'attack', 2, '{"damage":15,"draw_next_turn":2}',                     'uncommon', 'Deal 15 damage. Next turn draw 2 extra cards.'),
  ('Footwork',      'rogue',   'power',  1, '{"dexterity":2}',                                      'uncommon', 'Gain 2 Dexterity.'),
  ('Blade Dance',   'rogue',   'attack', 1, '{"damage":4,"hits":3}',                                'uncommon', 'Deal 4 damage 3 times.'),
  ('Prepared',      'rogue',   'skill',  0, '{"draw":1,"discard":1}',                               'common',   'Draw 1 card. Discard 1 card.'),

-- ─── Mage Cards ──────────────────────────────────────────────────────────────
  ('Zap',           'mage',    'attack', 1, '{"damage":7}',                                         'common',   'Deal 7 damage.'),
  ('Defend (Mage)', 'mage',    'skill',  1, '{"block":5}',                                          'common',   'Gain 5 Block.'),
  ('Ball Lightning','mage',    'attack', 1, '{"damage":7,"all_enemies":true}',                      'uncommon', 'Deal 7 damage to ALL enemies.'),
  ('Thunderclap',   'mage',    'attack', 1, '{"damage":4,"all_enemies":true,"apply_vulnerable":1}', 'common',   'Deal 4 damage to ALL enemies. Apply 1 Vulnerable.'),
  ('Glacier',       'mage',    'skill',  2, '{"block":8,"block_again":8}',                          'uncommon', 'Gain 8 Block twice.'),
  ('Dark Shackles', 'mage',    'skill',  0, '{"enemy_strength_this_turn":-9}',                      'uncommon', 'Enemy loses 9 Strength this turn.'),
  ('Cold Snap',     'mage',    'attack', 1, '{"damage":6,"gain_energy_if_kill":1}',                 'common',   'Deal 6 damage. If this kills an enemy, gain 1 Energy.'),
  ('Chaos Theory',  'mage',    'skill',  0, '{"play_random_cards":3}',                              'rare',     'Play the top 3 cards of your draw pile at random.'),
  ('Fission',       'mage',    'power',  0, '{"gain_energy_per_exhaust":1}',                        'rare',     'Whenever you Exhaust a card, gain 1 Energy.'),
  ('Aggregate',     'mage',    'skill',  1, '{"gain_energy_per_5_draw_pile":1}',                    'uncommon', 'Gain 1 Energy for every 5 cards in your draw pile.');

-- ─── Relic Seeds ─────────────────────────────────────────────────────────────
INSERT INTO deck.relic_definitions (name, rarity, effect, description) VALUES
  ('Burning Blood',      'starter',  '{"heal_after_combat":6}',                'Heal 6 HP after every combat.'),
  ('Ring of the Snake',  'starter',  '{"extra_draw_start":2}',                 'Start each combat by drawing 2 extra cards.'),
  ('Vajra',              'common',   '{"strength":1}',                         'Gain 1 Strength at start of each combat.'),
  ('Bag of Preparation', 'common',   '{"extra_draw_turn1":2}',                 'Draw 2 extra cards on turn 1 of each combat.'),
  ('Anchor',             'common',   '{"block_start_combat":10}',              'Start each combat with 10 Block.'),
  ('Red Skull',          'common',   '{"strength_when_low_hp":3,"hp_threshold":0.5}', 'When HP drops below 50%, gain 3 Strength.'),
  ('Philosopher''s Stone','uncommon', '{"energy_per_turn":1,"enemy_strength":1}','Gain 1 extra Energy each turn. Enemies start with 1 Strength.'),
  ('Bloody Idol',        'uncommon', '{"gold_after_combat":5}',                'Gain 5 Gold after each combat.'),
  ('Oddly Smooth Stone', 'uncommon', '{"dexterity_start":1}',                  'Start each combat with 1 Dexterity.'),
  ('Darkstone Periapt',  'uncommon', '{"max_hp":6}',                           'Increase Max HP by 6.');
