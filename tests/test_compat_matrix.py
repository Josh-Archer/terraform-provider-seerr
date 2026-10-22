import json
import unittest

from scripts import check_compat_matrix as m

COMPAT = {"seerr": {"release": "v3.5.0", "image": "seerr/seerr:v3.5.0"}}
WORKFLOW = "            image: seerr/seerr:v3.5.0\n  SEERR_TEST_IMAGE: ${{ vars.SEERR_TEST_IMAGE || 'seerr/seerr:v3.5.0' }}\n"
COMPOSE = "    image: ${SEERR_TEST_IMAGE:-seerr/seerr:v3.5.0}\n"
DOC = "Seerr `v3.5.0`"


class CompatMatrixTests(unittest.TestCase):
    def test_consistent(self):
        self.assertEqual(m.check(COMPAT, WORKFLOW, COMPOSE, DOC), [])

    def test_missing_matrix_image(self):
        self.assertTrue(m.check(COMPAT, WORKFLOW.replace("image: seerr", "image: other"), COMPOSE, DOC))

    def test_stale_default(self):
        self.assertTrue(m.check(COMPAT, WORKFLOW.replace("|| 'seerr/seerr:v3.5.0'", "|| 'seerr/seerr:v3.4.1'"), COMPOSE, DOC))
        self.assertTrue(m.check(COMPAT, WORKFLOW, COMPOSE.replace("3.5.0", "3.4.1"), DOC))

    def test_doc_drift(self):
        self.assertTrue(m.check(COMPAT, WORKFLOW, COMPOSE, "Seerr `v3.4.1`"))

    def test_repository_is_consistent(self):
        self.assertEqual(m.main(), 0)


if __name__ == "__main__":
    unittest.main()
