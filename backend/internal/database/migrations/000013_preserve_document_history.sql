CREATE FUNCTION preserve_document_history() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'Issued documents and print attempts are append-only';
END;
$$;
CREATE TRIGGER documents_preserve_history BEFORE UPDATE OR DELETE ON documents
FOR EACH ROW EXECUTE FUNCTION preserve_document_history();
CREATE TRIGGER document_print_attempts_preserve_history BEFORE UPDATE OR DELETE ON document_print_attempts
FOR EACH ROW EXECUTE FUNCTION preserve_document_history();
