<?php
class Reference extends CI_Controller {
    public function __construct() {
        parent::__construct();
        $this->load->database();
        $this->load->library('fixture_session');
    }
    private function reply($code, $body) {
        $this->output->set_status_header($code)->set_header('X-Fixture: ci3-source')
            ->set_content_type('application/json')->set_output(json_encode($body));
    }
    private function actor() {
        $s = $this->fixture_session->current();
        if (!$s) { $this->reply(401, ['error' => 'unauthenticated']); return null; }
        // Actual principal is independently correlated with DB/session snapshots.
        return $this->db->get_where('users', ['id' => $s['principal']])->row_array();
    }
    private function mutation() {
        $a = $this->actor(); if (!$a) { return null; }
        if (!$this->fixture_session->consume($this->input->post('csrf'))) {
            $this->reply(403, ['error' => 'csrf']); return null;
        }
        return $a;
    }
    public function health() { $this->reply(200, ['fixture' => 'ci3-3.1.13', 'db' => $this->db->query('SELECT 1 AS ready')->row()->ready]); }
    public function login() {
        $u = $this->db->get_where('users', ['name' => $this->input->post('name')])->row_array();
        if (!$u || !password_verify((string)$this->input->post('password'), $u['password'])) {
            $this->reply(401, ['error' => 'credentials']); return;
        }
        $this->fixture_session->login($u['id']);
        $this->output->set_header('Location: /whoami');
        $this->reply(303, ['principal' => (int)$u['id']]);
    }
    public function whoami() {
        $a = $this->actor(); if (!$a) { return; }
        $this->reply(200, ['principal' => (int)$a['id'], 'name' => $a['name'], 'role' => $a['role'], 'flash' => $this->fixture_session->flash()]);
    }
    public function csrf() {
        if (!$this->actor()) { return; }
        $this->reply(200, ['token' => $this->fixture_session->issue()]);
    }
    public function create() {
        $a = $this->mutation(); if (!$a) { return; }
        $this->db->trans_start();
        $this->db->insert('items', ['owner' => $a['id'], 'name' => (string)$this->input->post('name')]);
        $id = $this->db->insert_id();
        $this->db->insert('jobs', ['item' => $id, 'status' => 'pending']);
        // Disposable qualification metadata; never authentication/expected effects.
        $capture = json_decode((string)$this->input->post('_rwt_capture'), true);
        if (is_array($capture)) {
            $job = $this->db->insert_id();
            $capture['job_id'] = 'job_'.$job;
            $capture['effect_id'] = 'effect_'.$job;
            $this->db->where('id', $job)->update('jobs', ['capture_context' => json_encode($capture)]);
        }
        $this->db->trans_complete();
        $this->reply(202, ['id' => (int)$id, 'owner' => (int)$a['id']]);
    }
    public function item() {
        $a = $this->actor(); if (!$a) { return; }
        // Direct PDO deliberately bypasses CI3 driver instrumentation (X11/X70).
        $pdo = new PDO('mysql:host=db;dbname=fixture;charset=utf8mb4', 'fixture', 'synthetic-fixture-only');
        $q = $pdo->prepare('SELECT id,owner,name FROM items WHERE id=?');
        $q->execute([$this->input->get('id')]); $i = $q->fetch(PDO::FETCH_ASSOC);
        if (!$i || (int)$i['owner'] !== (int)$a['id']) { $this->reply(403, ['error' => 'owner']); return; }
        $this->reply(200, ['id' => (int)$i['id'], 'owner' => (int)$i['owner'], 'name' => $i['name']]);
    }
    public function note() {
        if (!$this->mutation()) { return; }
        file_put_contents('/state/notes/note.txt', (string)$this->input->post('text'), LOCK_EX);
        chmod('/state/notes/note.txt', 0640);
        $this->reply(200, ['saved' => true]);
    }
    public function logout() {
        if (!$this->mutation()) { return; }
        $this->fixture_session->invalidate();
        header('Set-Cookie: fixture_session=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax', false);
        $this->reply(200, ['logout' => true]);
    }
}
