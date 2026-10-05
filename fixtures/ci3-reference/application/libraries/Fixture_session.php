<?php
// Explicit safe JSON store: opaque IDs, no unserialize/eval or ambient secrets.
class Fixture_session {
    private $id;
    private $state;
    public function __construct() {
        $this->id = isset($_COOKIE['fixture_session']) ? $_COOKIE['fixture_session'] : '';
        $this->state = null;
        if (preg_match('/^[a-f0-9]{64}$/D', $this->id)) {
            $path = '/state/sessions/'.$this->id.'.json';
            if (is_file($path)) {
                $value = json_decode(file_get_contents($path), true);
                if (is_array($value) && $value['expires'] > time()) { $this->state = $value; }
            }
        }
    }
    public function current() { return $this->state; }
    public function login($principal) {
        $this->invalidate();
        $this->id = bin2hex(random_bytes(32));
        $this->state = ['principal' => (int)$principal, 'expires' => time()+60,
            'csrf' => null, 'csrf_expires' => null, 'flash' => 'logged-in'];
        $this->save();
        header('Set-Cookie: fixture_session='.$this->id.'; Path=/; Max-Age=60; HttpOnly; SameSite=Lax', false);
        header('Set-Cookie: fixture_marker=source; Path=/; SameSite=Strict', false);
    }
    public function issue() {
        $this->state['csrf'] = bin2hex(random_bytes(32));
        $this->state['csrf_expires'] = time()+10;
        $this->save();
        return $this->state['csrf'];
    }
    public function consume($token) {
        if (!$this->state || !is_string($token) || !$this->state['csrf'] ||
            $this->state['csrf_expires'] <= time() || !hash_equals($this->state['csrf'], $token)) { return false; }
        $this->state['csrf'] = null;
        $this->state['csrf_expires'] = null;
        $this->save();
        return true;
    }
    public function flash() {
        $flash = $this->state['flash']; $this->state['flash'] = null; $this->save(); return $flash;
    }
    public function invalidate() {
        if (preg_match('/^[a-f0-9]{64}$/D', $this->id)) {
            $p = '/state/sessions/'.$this->id.'.json'; if (is_file($p)) { unlink($p); }
        }
        $this->state = null;
    }
    private function save() {
        file_put_contents('/state/sessions/'.$this->id.'.json', json_encode($this->state), LOCK_EX);
        chmod('/state/sessions/'.$this->id.'.json', 0600);
    }
}
