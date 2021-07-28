;;; cozy.el --- mode for editing cozy scripts

;; Copyright (C) 2018 Steve Kemp

;; Author: Steve Kemp <steve@steve.fi>
;; Keywords: languages
;; Version: 1.0

;;; Commentary:

;; Provides support for editing cozy scripts with full support for
;; font-locking, but no special keybindings, or indentation handling.

;;;; Enabling:

;; Add the following to your .emacs file

;; (require 'cozy)
;; (setq auto-mode-alist (append '(("\\.mon$" . cozy-mode)) auto-mode-alist)))



;;; Code:

(defvar cozy-constants
  '("true"
    "false"))

(defvar cozy-keywords
  '(
    "else"
    "fn"
    "for"
    "foreach"
    "if"
    "in"
    "let"
    "mutable"
    "return"
    ))

;; The language-core and functions from the standard-library.
(defvar cozy-functions
  '(
    "async"
    "await"
    "error"
    "exit"
    "file"
    "first"
    "import"
    "int"
    "last"
    "len"
    "macro"
    "match"
    "math"
    "print"
    "quote"
    "read"
    "rest"
    "string"
    "type"
    "unquote"
    "version"
    ))


(defvar cozy-font-lock-defaults
  `((
     ("\"\\.\\*\\?" . font-lock-string-face)
     (";\\|,\\|=" . font-lock-keyword-face)
     ( ,(regexp-opt cozy-keywords 'words) . font-lock-builtin-face)
     ( ,(regexp-opt cozy-constants 'words) . font-lock-constant-face)
     ( ,(regexp-opt cozy-functions 'words) . font-lock-function-name-face)
     )))

(define-derived-mode cozy-mode fundamental-mode "cozy script"
  "cozy-mode is a major mode for editing cozy scripts"
  (setq font-lock-defaults cozy-font-lock-defaults)

  ;; Comment handler for single & multi-line modes
  (modify-syntax-entry ?\/ ". 124b" cozy-mode-syntax-table)
  (modify-syntax-entry ?\* ". 23n" cozy-mode-syntax-table)

  ;; Comment ender for single-line comments.
  (modify-syntax-entry ?\n "> b" cozy-mode-syntax-table)
  (modify-syntax-entry ?\r "> b" cozy-mode-syntax-table)
  )

(provide 'cozy)
