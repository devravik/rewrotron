CREATE DATABASE fixture CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
CREATE USER 'fixture'@'%' IDENTIFIED BY 'synthetic-fixture-only';
GRANT SELECT,INSERT,UPDATE,DELETE ON fixture.* TO 'fixture'@'%';
CREATE USER 'observer'@'%' IDENTIFIED BY 'synthetic-observer-only';
GRANT SELECT ON fixture.* TO 'observer'@'%';
USE fixture;
CREATE TABLE users(id INT PRIMARY KEY,name VARCHAR(20) UNIQUE,password VARCHAR(100),role VARCHAR(20));
INSERT INTO users VALUES
(1,'alice','$2y$10$abcdefghijklmnopqrstuu8vlHjBPcS/rvADQ/wyGPn4IRDgLYpvW','owner'),
(2,'bob','$2y$10$abcdefghijklmnopqrstuu8vlHjBPcS/rvADQ/wyGPn4IRDgLYpvW','other');
CREATE TABLE items(id INT AUTO_INCREMENT PRIMARY KEY,owner INT NOT NULL,name VARCHAR(100),FOREIGN KEY(owner) REFERENCES users(id));
CREATE TABLE audit(id INT AUTO_INCREMENT PRIMARY KEY,item INT,owner INT);
CREATE TRIGGER item_audit AFTER INSERT ON items FOR EACH ROW INSERT INTO audit(item,owner) VALUES(NEW.id,NEW.owner);
CREATE TABLE jobs(id INT AUTO_INCREMENT PRIMARY KEY,item INT,status VARCHAR(20),receipt VARCHAR(64),capture_context TEXT);
