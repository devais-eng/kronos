from contextlib import contextmanager


@contextmanager
def read_session(close=True, **kwargs):
    from kronos.storage import SessionFactory
    session = SessionFactory(**kwargs)
    try:
        yield session
    finally:
        # Close the session if desired
        if close:
            session.close()


@contextmanager
def transactional_session(close=False, **kwargs):
    from kronos.storage import SessionFactory
    session = SessionFactory(**kwargs)
    raised_exception = None
    try:
        yield session
    except Exception as e:  # noqa
        # Roll back if the nested block raised an error
        session.rollback()
        raised_exception = e
    else:
        # Commit if it didn't (so flow ran off the end of the try block)
        session.commit()
    finally:
        # Close the session if desired
        if close:
            session.close()
        if raised_exception is not None:
            raise raised_exception
